package action

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

func PromiseOwner(snapshot Snapshot, promise string) (string, bool) {
	for _, capability := range snapshot.Capabilities {
		for _, item := range capability.Promises() {
			if item.ID().String() == promise {
				return capability.ID().String(), true
			}
		}
	}

	return "", false
}

func OpenSubtasks(snapshot Snapshot, task string) []string {
	prefix := task + "."
	var open []string

	for _, candidate := range snapshot.Tasks {
		id := candidate.ID().String()
		if !strings.HasPrefix(id, prefix) {
			continue
		}

		if candidate.Status() != agenda.TaskDone {
			open = append(open, id)
		}
	}
	slices.Sort(open)

	return open
}

type ClosedRecord struct {
	Kind thread.NodeKind
	ID   string
}

func InterruptedCloses(
	snapshot Snapshot,
	closed []ClosedRecord,
) []ClosedRecord {
	active := activeKinds(snapshot)
	var interrupted []ClosedRecord

	for _, record := range closed {
		if kind, found := active[record.ID]; found && kind == record.Kind {
			interrupted = append(interrupted, record)
		}
	}

	return interrupted
}

func activeKinds(snapshot Snapshot) map[string]thread.NodeKind {
	kinds := make(map[string]thread.NodeKind)
	mark := func(id fmt.Stringer, kind thread.NodeKind) {
		kinds[id.String()] = kind
	}

	for _, goal := range snapshot.Goals {
		mark(goal.ID(), thread.NodeGoal)
	}
	for _, capability := range snapshot.Capabilities {
		mark(capability.ID(), thread.NodeCapability)
		for _, promise := range capability.Promises() {
			mark(promise.ID(), thread.NodePromise)
		}
	}
	for _, decision := range snapshot.Decisions {
		mark(decision.ID(), thread.NodeDecision)
	}
	for _, question := range snapshot.Questions {
		mark(question.ID(), thread.NodeQuestion)
	}
	for _, work := range snapshot.Works {
		mark(work.ID(), thread.NodeWork)
	}
	for _, task := range snapshot.Tasks {
		mark(task.ID(), thread.NodeTask)
	}

	return kinds
}

func CompletedCloseFindings(closed []ClosedRecord) []Finding {
	findings := make([]Finding, 0, len(closed))

	for _, record := range closed {
		findings = append(findings, Finding{
			Severity: FindingWarning,
			Code:     CodeCloseCompleted,
			Message:  closeCompletedMessage,
			Details:  map[string]any{"id": record.ID},
		})
	}

	return findings
}
