package action

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

type ReadyResult struct {
	Works []agenda.Work
	Tasks []agenda.Task
}

type readiness struct {
	blocked     map[kernel.RecordID]bool
	openWorks   map[kernel.RecordID]bool
	openParents map[kernel.RecordID]bool
	claims      map[kernel.RecordID]agenda.TaskClaim
	branch      agenda.TaskBranch
}

func Ready(
	ctx context.Context,
	repositories Repositories,
	claims agenda.ClaimRepository,
	branch agenda.TaskBranch,
) (ReadyResult, []error) {
	snapshot := Load(ctx, repositories)
	graph, _ := Validate(snapshot)
	failures := snapshot.Failures

	local, err := claims.ListClaims(ctx)
	if err != nil {
		failures = append(failures, err)
	}

	return readyWork(snapshot, graph, local, branch), failures
}

func readyWork(
	snapshot Snapshot,
	graph thread.Graph,
	local []agenda.TaskClaim,
	branch agenda.TaskBranch,
) ReadyResult {
	state := readiness{
		blocked:     blockedTargets(graph),
		openWorks:   make(map[kernel.RecordID]bool),
		openParents: openParents(snapshot.Tasks),
		claims:      claimsByTask(local),
		branch:      branch,
	}

	var result ReadyResult
	for _, work := range snapshot.Works {
		id, err := recordIDOf(work.ID())
		if err != nil {
			continue
		}

		if work.Status() == agenda.WorkActive && !state.blocked[id] {
			state.openWorks[id] = true
			result.Works = append(result.Works, work)
		}
	}

	for _, task := range snapshot.Tasks {
		if state.allows(task) {
			result.Tasks = append(result.Tasks, task)
		}
	}

	return result
}

func blockedTargets(graph thread.Graph) map[kernel.RecordID]bool {
	blocked := make(map[kernel.RecordID]bool)

	for _, edge := range thread.ActiveBlocks(graph) {
		blocked[edge.Target] = true
	}

	return blocked
}

func openParents(tasks []agenda.Task) map[kernel.RecordID]bool {
	parents := make(map[kernel.RecordID]bool)

	for _, task := range tasks {
		id, err := recordIDOf(task.ID())
		if err != nil || task.Status() == agenda.TaskDone {
			continue
		}

		if parent, found := id.Parent(); found {
			parents[parent] = true
		}
	}

	return parents
}

func (r readiness) allows(task agenda.Task) bool {
	id, err := recordIDOf(task.ID())

	switch {
	case err != nil:
		return false
	case !r.isAvailable(task, id):
		return false
	case !r.openWorks[id.Root()]:
		return false
	case r.openParents[id]:
		return false
	}

	return !r.isBlocked(id)
}

func claimsByTask(
	claims []agenda.TaskClaim,
) map[kernel.RecordID]agenda.TaskClaim {
	index := make(map[kernel.RecordID]agenda.TaskClaim, len(claims))

	for _, claim := range claims {
		if id, err := recordIDOf(claim.Task); err == nil {
			index[id] = claim
		}
	}

	return index
}

func (r readiness) isAvailable(task agenda.Task, id kernel.RecordID) bool {
	if task.Status() == agenda.TaskDone {
		return false
	}

	if claim, found := r.claims[id]; found {
		return claim.Branch == r.branch
	}

	switch task.Status() {
	case agenda.TaskOpen:
		return true
	case agenda.TaskClaimed:
		branch, found := task.Branch()
		return found && branch == r.branch
	}

	return false
}

func (r readiness) isBlocked(id kernel.RecordID) bool {
	root := id.Root()

	for current := id; current != root; {
		if r.blocked[current] {
			return true
		}

		parent, found := current.Parent()
		if !found {
			return false
		}

		current = parent
	}

	return false
}
