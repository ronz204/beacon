package action

import (
	"fmt"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
)

func IntroducedFailures(snapshot Snapshot, entity any) []error {
	_, before := Validate(snapshot)
	_, after := Validate(withRecord(snapshot, entity))

	known := make(map[string]bool, len(before))
	for _, failure := range before {
		known[failureKey(failure)] = true
	}

	return slices.DeleteFunc(after, func(failure error) bool {
		return known[failureKey(failure)]
	})
}

func failureKey(failure error) string {
	return fmt.Sprintf("%+v", kernel.DescribeFailure(failure))
}

func RecordIDs(snapshot Snapshot) map[string]bool {
	ids := make(map[string]bool)
	add := func(id fmt.Stringer) { ids[id.String()] = true }

	for _, goal := range snapshot.Goals {
		add(goal.ID())
	}
	for _, capability := range snapshot.Capabilities {
		add(capability.ID())
		for _, promise := range capability.Promises() {
			add(promise.ID())
		}
	}
	for _, decision := range snapshot.Decisions {
		add(decision.ID())
	}
	for _, question := range snapshot.Questions {
		add(question.ID())
	}
	for _, guide := range snapshot.Guides {
		add(guide.ID())
	}
	for _, work := range snapshot.Works {
		add(work.ID())
	}
	for _, task := range snapshot.Tasks {
		add(task.ID())
	}
	for _, node := range snapshot.Archived {
		add(node.ID)
	}

	return ids
}

func withRecord(snapshot Snapshot, entity any) Snapshot {
	switch value := entity.(type) {
	case ledger.Goal:
		snapshot.Goals = replaced(snapshot.Goals, value, ledger.Goal.ID)
	case ledger.Capability:
		snapshot.Capabilities = replaced(
			snapshot.Capabilities,
			value,
			ledger.Capability.ID,
		)
	case memory.Decision:
		snapshot.Decisions = replaced(
			snapshot.Decisions,
			value,
			memory.Decision.ID,
		)
	case memory.Question:
		snapshot.Questions = replaced(
			snapshot.Questions,
			value,
			memory.Question.ID,
		)
	case memory.Guide:
		snapshot.Guides = replaced(snapshot.Guides, value, memory.Guide.ID)
	case agenda.Work:
		snapshot.Works = replaced(snapshot.Works, value, agenda.Work.ID)
	case agenda.Task:
		snapshot.Tasks = replaced(snapshot.Tasks, value, agenda.Task.ID)
	}

	return snapshot
}

func replaced[T any, I fmt.Stringer](items []T, item T, id func(T) I) []T {
	target := id(item).String()
	result := slices.Clone(items)

	for index := range result {
		if id(result[index]).String() == target {
			result[index] = item
			return result
		}
	}

	return append(result, item)
}
