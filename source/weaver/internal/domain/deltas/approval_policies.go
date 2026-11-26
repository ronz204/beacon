package deltas

import (
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

func StampApproval(
	actor kernel.Actor,
	human string,
	today kernel.CalendarDate,
) (kernel.ApprovalInput, error) {
	if actor.IsHuman() {
		return kernel.ApprovalInput{
			By: actor.String(),
			At: today.String(),
		}, nil
	}

	name := strings.TrimSpace(human)
	if name == "" {
		return kernel.ApprovalInput{}, approvalErrors.Fail(
			ApprovalIdentityMissing,
			nil,
		)
	}

	return kernel.ApprovalInput{
		By:  "human:" + name,
		At:  today.String(),
		Via: actor.String(),
	}, nil
}

func IsFreshlyApproved(target ChangeTarget, changes []StateChange) bool {
	if isApprovedClosure(target, changes) {
		return true
	}

	switch target.kind {
	case TargetPromise:
		return freshPromiseApproval(target.value, changes)
	case TargetConfig:
		return freshRecordApproval(isConfigRecord, changes)
	case TargetRecord:
		return freshRecordApproval(recordNamed(target.value), changes)
	case TargetBaseline:
		return freshBaselineApproval(target.value, changes)
	}

	return false
}

func isApprovedClosure(target ChangeTarget, changes []StateChange) bool {
	for _, change := range changes {
		for _, closed := range change.closed {
			sameID := closed.id.String() == target.value
			if sameID && closed.approved != "" {
				return true
			}
		}
	}

	return false
}

func freshRecordApproval(
	selected func(*ChangedRecord) bool,
	changes []StateChange,
) bool {
	for _, change := range changes {
		record := change.after
		if record == nil || change.baseline || !selected(record) {
			continue
		}

		if isNewApproval(change.before, record) {
			return true
		}
	}

	return false
}

func isNewApproval(before, after *ChangedRecord) bool {
	if after.approved == "" {
		return false
	}

	return before == nil || before.approved != after.approved
}

func freshPromiseApproval(id string, changes []StateChange) bool {
	return anyPromiseApproval(changes, func(p ChangedPromise) bool {
		return p.id.String() == id
	})
}

func freshBaselineApproval(path string, changes []StateChange) bool {
	runners := baselineRunners(path, changes)

	return anyPromiseApproval(changes, func(p ChangedPromise) bool {
		return p.runner != "" && slices.Contains(runners, p.runner)
	})
}

func baselineRunners(path string, changes []StateChange) []string {
	for _, change := range changes {
		if change.baseline && change.path == path {
			return change.runners
		}
	}

	return nil
}

func anyPromiseApproval(
	changes []StateChange,
	selected func(ChangedPromise) bool,
) bool {
	for _, change := range changes {
		if change.after == nil || change.baseline {
			continue
		}

		for _, next := range change.after.promises {
			if selected(next) && newPromiseApproval(change.before, next) {
				return true
			}
		}
	}

	return false
}

func newPromiseApproval(before *ChangedRecord, next ChangedPromise) bool {
	if next.approved == "" {
		return false
	}
	if before == nil {
		return true
	}

	previous, existed := before.promiseByID(next.id)

	return !existed || previous.approved != next.approved
}

func isConfigRecord(record *ChangedRecord) bool {
	return record.kind == RecordConfig
}

func recordNamed(id string) func(*ChangedRecord) bool {
	return func(record *ChangedRecord) bool {
		return record.kind != RecordConfig && record.id.String() == id
	}
}
