package action

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

type Repositories struct {
	Goals        ledger.GoalRepository
	Capabilities ledger.CapabilityRepository
	Decisions    memory.DecisionRepository
	Questions    memory.QuestionRepository
	Guides       memory.GuideRepository
	Works        agenda.WorkRepository
	Tasks        agenda.TaskRepository
	Archive      thread.ArchiveIndex
}

type Snapshot struct {
	Goals        []ledger.Goal
	Capabilities []ledger.Capability
	Decisions    []memory.Decision
	Questions    []memory.Question
	Guides       []memory.Guide
	Works        []agenda.Work
	Tasks        []agenda.Task
	Archived     []thread.GraphNode
	Failures     []error
}

func Load(ctx context.Context, repositories Repositories) Snapshot {
	var snapshot Snapshot
	var failures []error

	snapshot.Goals, failures = repositories.Goals.ListGoals(ctx)
	snapshot.Failures = append(snapshot.Failures, failures...)

	snapshot.Capabilities, failures =
		repositories.Capabilities.ListCapabilities(ctx)
	snapshot.Failures = append(snapshot.Failures, failures...)

	snapshot.Decisions, failures = repositories.Decisions.ListDecisions(ctx)
	snapshot.Failures = append(snapshot.Failures, failures...)

	snapshot.Questions, failures = repositories.Questions.ListQuestions(ctx)
	snapshot.Failures = append(snapshot.Failures, failures...)

	snapshot.Guides, failures = repositories.Guides.ListGuides(ctx)
	snapshot.Failures = append(snapshot.Failures, failures...)

	snapshot.Works, failures = repositories.Works.ListWorks(ctx)
	snapshot.Failures = append(snapshot.Failures, failures...)

	snapshot.Tasks, failures = repositories.Tasks.ListTasks(ctx)
	snapshot.Failures = append(snapshot.Failures, failures...)

	return withArchive(ctx, snapshot, repositories.Archive)
}

func withArchive(
	ctx context.Context,
	snapshot Snapshot,
	archive thread.ArchiveIndex,
) Snapshot {
	if archive == nil || !missesRecords(snapshot) {
		return snapshot
	}

	nodes, err := archive.ArchivedNodes(ctx)
	if err != nil {
		snapshot.Failures = append(snapshot.Failures, err)
		return snapshot
	}

	snapshot.Archived = nodes

	return snapshot
}

func missesRecords(snapshot Snapshot) bool {
	_, failures := Validate(snapshot)

	for _, failure := range failures {
		code := kernel.DescribeFailure(failure).Code
		if code == string(thread.GraphTargetMissing) ||
			code == string(thread.GraphTaskOrphan) {
			return true
		}
	}

	return false
}
