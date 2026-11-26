package deltas

import (
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/source/kernel"
)

type Finding struct {
	Target   ChangeTarget
	Level    ChangeLevel
	NeedsYes bool
	Reason   ChangeReason
	Before   string
	After    string
}

func (f Finding) because(
	reason ChangeReason,
	before, after string,
) Finding {
	f.Reason = reason
	f.Before = before
	f.After = after

	return f
}

var goalTextFields = []string{"statement", "success"}

var configGuardedFields = []string{"runners", "schema"}

var configReasons = map[string]ChangeReason{
	"runners": ReasonRunnersChanged,
	"schema":  ReasonSchemaChanged,
}

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
	if record == nil {
		return nil
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

	finding := weakening(target).because(ReasonBaselineChanged, "", "")

	return []Finding{finding}
}

func configFindings(change StateChange) []Finding {
	target := ConfigTarget()

	for _, field := range configGuardedFields {
		before, after := fieldPair(change, field)

		if before != after {
			finding := weakening(target).because(configReasons[field], "", "")

			return []Finding{finding}
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
	statement, added := fieldPair(change, "statement")

	if change.before == nil {
		finding := weakening(target).because(ReasonGoalAdded, "", added)

		return []Finding{finding}
	}
	if change.after == nil {
		finding := weakening(target).because(ReasonGoalClosed, statement, "")

		return []Finding{finding}
	}

	for _, field := range goalTextFields {
		before, after := fieldPair(change, field)

		if NormalizeText(before) != NormalizeText(after) {
			return []Finding{goalChanged(target, before, after)}
		}
	}

	before, after := fieldPair(change, "status")
	if before != after {
		return []Finding{goalChanged(target, before, after)}
	}

	return []Finding{maintenance(target)}
}

func goalChanged(target ChangeTarget, before, after string) Finding {
	return weakening(target).because(ReasonGoalChanged, before, after)
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

	var findings []Finding

	switch {
	case change.before == nil:
		findings = append(findings, addition(recordTarget(record), false))
	case !maps.Equal(change.before.fields, record.fields):
		findings = append(findings, maintenance(recordTarget(record)))
	}

	return append(findings, promiseFindings(change)...)
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
		if existed && previous.sameContent(next) {
			continue
		}

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
	case !existed && next.active:
		return addition(target, true).because(
			ReasonPromiseNewActive,
			"",
			next.text,
		)
	case !existed:
		return addition(target, false)
	case previous.active && !next.active:
		return weakening(target).because(
			ReasonPromiseDemoted,
			previous.text,
			"",
		)
	case previous.active:
		return activePromiseFinding(previous, next)
	case next.active:
		return addition(target, true).because(
			ReasonPromiseActivated,
			"",
			next.text,
		)
	}

	return maintenance(target)
}

func activePromiseFinding(previous, next ChangedPromise) Finding {
	target := PromiseTarget(next.id)

	if previous.check != next.check {
		return weakening(target).because(
			ReasonPromiseCheck,
			previous.check,
			next.check,
		)
	}
	if NormalizeText(previous.text) != NormalizeText(next.text) {
		return weakening(target).because(
			ReasonPromiseText,
			previous.text,
			next.text,
		)
	}

	return maintenance(target)
}

func removedPromiseFinding(previous ChangedPromise) Finding {
	target := PromiseTarget(previous.id)

	if previous.active {
		return weakening(target).because(
			ReasonPromiseClosed,
			previous.text,
			"",
		)
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
