package action

import (
	"context"

	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
)

type Repositories struct {
	Goals        ledger.GoalRepository
	Capabilities ledger.CapabilityRepository
	Decisions    memory.DecisionRepository
	Questions    memory.QuestionRepository
	Guides       memory.GuideRepository
	Works        agenda.WorkRepository
	Tasks        agenda.TaskRepository
}

type Snapshot struct {
	Goals        []ledger.Goal
	Capabilities []ledger.Capability
	Decisions    []memory.Decision
	Questions    []memory.Question
	Guides       []memory.Guide
	Works        []agenda.Work
	Tasks        []agenda.Task
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

	return snapshot
}
