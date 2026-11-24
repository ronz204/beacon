package action

import (
	"context"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/internal/spider/internal/domain/thread"
)

type ReadyResult struct {
	Works []agenda.Work
	Tasks []agenda.Task
}

type readiness struct {
	blocked     map[kernel.RecordID]bool
	openWorks   map[kernel.RecordID]bool
	openParents map[kernel.RecordID]bool
	branch      agenda.TaskBranch
}

func Ready(
	ctx context.Context,
	repositories Repositories,
	branch agenda.TaskBranch,
) (ReadyResult, []error) {
	snapshot := Load(ctx, repositories)
	graph, _ := Validate(snapshot)

	return readyWork(snapshot, graph, branch), snapshot.Failures
}

func readyWork(
	snapshot Snapshot,
	graph thread.Graph,
	branch agenda.TaskBranch,
) ReadyResult {
	state := readiness{
		blocked:     blockedTargets(graph),
		openWorks:   make(map[kernel.RecordID]bool),
		openParents: openParents(snapshot.Tasks),
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
	case !r.isAvailable(task):
		return false
	case !r.openWorks[id.Root()]:
		return false
	case r.openParents[id]:
		return false
	}

	return !r.isBlocked(id)
}

func (r readiness) isAvailable(task agenda.Task) bool {
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
