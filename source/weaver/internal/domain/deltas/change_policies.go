package deltas

import (
	"slices"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/source/kernel"
)

type Finding struct {
	Target   ChangeTarget
	Level    ChangeLevel
	NeedsYes bool
}

var goalTextFields = []string{"statement", "success"}

var configGuardedFields = []string{"runners", "schema"}

func NormalizeText(text string) string {
	lowered := strings.ToLower(text)
	spaced := strings.Map(replacePunctuation, lowered)

	return strings.Join(strings.Fields(spaced), " ")
}

func replacePunctuation(r rune) rune {
	if unicode.IsPunct(r) {
		return ' '
	}

	return r
}

func ClassifyChange(change StateChange) (ChangeLevel, bool) {
	level := LevelMaintenance
	needsYes := false

	for _, finding := range ClassifyElements(change) {
		level = max(level, finding.Level)
		needsYes = needsYes || finding.NeedsYes
	}

	return level, needsYes
}

func ClassifyElements(change StateChange) []Finding {
	if change.baseline {
		return baselineFindings(change)
	}

	record := change.after
	if record == nil {
		record = change.before
	}

	switch record.kind {
	case RecordConfig:
		return configFindings(change)
	case RecordGoal:
		return goalFindings(change)
	case RecordCapability:
		return capabilityFindings(change)
	case RecordDecision:
		return decisionFindings(change)
	}

	return []Finding{maintenance(recordTarget(record))}
}

func baselineFindings(change StateChange) []Finding {
	target, err := BaselineTarget(change.path)
	if err != nil {
		return nil
	}

	return []Finding{{Target: target, Level: LevelWeakening, NeedsYes: true}}
}

func configFindings(change StateChange) []Finding {
	target := ConfigTarget()

	for _, field := range configGuardedFields {
		before, after := fieldPair(change, field)

		if before != after {
			return []Finding{weakening(target)}
		}
	}

	return []Finding{maintenance(target)}
}

func goalFindings(change StateChange) []Finding {
	record := change.after
	if record == nil {
		record = change.before
	}

	target := recordTarget(record)
	if change.before == nil || change.after == nil {
		return []Finding{weakening(target)}
	}

	for _, field := range goalTextFields {
		before, after := fieldPair(change, field)

		if NormalizeText(before) != NormalizeText(after) {
			return []Finding{weakening(target)}
		}
	}

	before, after := fieldPair(change, "status")
	if before != after {
		return []Finding{weakening(target)}
	}

	return []Finding{maintenance(target)}
}

func decisionFindings(change StateChange) []Finding {
	if change.kind == ChangeAdded {
		return []Finding{addition(recordTarget(change.after), false)}
	}

	record := change.after
	if record == nil {
		record = change.before
	}

	return []Finding{maintenance(recordTarget(record))}
}

func capabilityFindings(change StateChange) []Finding {
	record := change.after
	if record == nil {
		record = change.before
	}

	findings := []Finding{capabilityOwnFinding(change, record)}
	findings = append(findings, promiseFindings(change)...)

	return findings
}

func capabilityOwnFinding(change StateChange, record *ChangedRecord) Finding {
	target := recordTarget(record)

	if change.before == nil {
		return addition(target, false)
	}

	return maintenance(target)
}

func promiseFindings(change StateChange) []Finding {
	var findings []Finding

	var before, after []ChangedPromise
	if change.before != nil {
		before = change.before.promises
	}
	if change.after != nil {
		after = change.after.promises
	}

	for _, next := range after {
		previous, existed := promiseIn(before, next.id)

		findings = append(findings, promiseFinding(previous, next, existed))
	}
	for _, previous := range before {
		if _, kept := promiseIn(after, previous.id); !kept {
			findings = append(findings, removedPromiseFinding(previous))
		}
	}

	return findings
}

func promiseIn(
	promises []ChangedPromise,
	id kernel.RecordID,
) (ChangedPromise, bool) {
	index := slices.IndexFunc(promises, func(p ChangedPromise) bool {
		return p.id == id
	})
	if index < 0 {
		return ChangedPromise{}, false
	}

	return promises[index], true
}

func promiseFinding(previous, next ChangedPromise, existed bool) Finding {
	target := PromiseTarget(next.id)

	switch {
	case !existed:
		return addition(target, next.active)
	case previous.active && !next.active:
		return weakening(target)
	case previous.active && promiseRewritten(previous, next):
		return weakening(target)
	case !previous.active && next.active:
		return addition(target, true)
	}

	return maintenance(target)
}

func promiseRewritten(previous, next ChangedPromise) bool {
	if previous.check != next.check {
		return true
	}

	return NormalizeText(previous.text) != NormalizeText(next.text)
}

func removedPromiseFinding(previous ChangedPromise) Finding {
	target := PromiseTarget(previous.id)

	if previous.active {
		return weakening(target)
	}

	return maintenance(target)
}

func fieldPair(change StateChange, field string) (string, string) {
	var before, after string

	if change.before != nil {
		before = change.before.fields[field]
	}
	if change.after != nil {
		after = change.after.fields[field]
	}

	return before, after
}

func recordTarget(record *ChangedRecord) ChangeTarget {
	return RecordTarget(record.id)
}

func maintenance(target ChangeTarget) Finding {
	return Finding{Target: target, Level: LevelMaintenance}
}

func addition(target ChangeTarget, needsYes bool) Finding {
	return Finding{Target: target, Level: LevelAddition, NeedsYes: needsYes}
}

func weakening(target ChangeTarget) Finding {
	return Finding{Target: target, Level: LevelWeakening, NeedsYes: true}
}
