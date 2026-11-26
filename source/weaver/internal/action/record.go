package action

import (
	"context"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type RecordCode string

const (
	ApprovalRequired RecordCode = "approval_required"
)

type RecordError = kernel.Error[RecordCode]

var recordErrors = kernel.ErrorCatalog[RecordCode]{
	ApprovalRequired: "some changes need a yes from the person",
}

type PreparedRecord struct {
	Type      string
	ID        string
	File      string
	Change    deltas.StateChange
	Ignored   []string
	Unchanged bool
	Token     any
}

type RecordStore interface {
	Prepare(
		ctx context.Context,
		payload []byte,
		author string,
	) (PreparedRecord, error)
	Save(
		ctx context.Context,
		prepared PreparedRecord,
		stamps map[string]kernel.ApprovalInput,
	) error
}

type Identity interface {
	Whoami(ctx context.Context) (Person, error)
}

type Person struct {
	Actor kernel.Actor
	Human string
}

type RecordInput struct {
	Payload []byte
	Approve bool
}

type ChangeReport struct {
	Target   string
	ID       string
	Level    int
	NeedsYes bool
	Reason   string
	Before   string
	After    string
}

type RecordResult struct {
	Type      string
	ID        string
	File      string
	Saved     bool
	Unchanged bool
	Changes   []ChangeReport
	Ignored   []string
}

type PendingSource interface {
	Pending(ctx context.Context) ([]Entry, error)
}

type Recorder struct {
	Store    RecordStore
	Identity Identity
	Today    kernel.CalendarDate
	Pending  PendingSource
}

func (r Recorder) Record(
	ctx context.Context,
	input RecordInput,
) (RecordResult, error) {
	person, err := r.Identity.Whoami(ctx)
	if err != nil {
		return RecordResult{}, err
	}

	prepared, err := r.Store.Prepare(
		ctx,
		input.Payload,
		person.Actor.String(),
	)
	if err != nil {
		return RecordResult{}, err
	}

	findings := deltas.ClassifyElements(prepared.Change)
	result := resultOf(prepared, findings)

	asking := askingFindings(findings)
	keys := keysOf(asking)

	if prepared.Unchanged {
		keys, result.Changes, err = r.approvalsOnDisk(
			ctx,
			prepared,
			input.Approve,
		)
		if err != nil || len(keys) == 0 {
			return result, err
		}
	}
	if len(keys) > 0 && !input.Approve {
		return result, approvalRequired(asking)
	}

	stamps, err := r.stamps(person, keys)
	if err != nil {
		return result, err
	}

	if err := r.Store.Save(ctx, prepared, stamps); err != nil {
		return result, err
	}

	result.Saved = true

	return result, nil
}

func (r Recorder) approvalsOnDisk(
	ctx context.Context,
	prepared PreparedRecord,
	approve bool,
) ([]string, []ChangeReport, error) {
	if !approve || r.Pending == nil {
		return nil, nil, nil
	}

	pending, err := r.Pending.Pending(ctx)
	if IsBaseMissing(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	var keys []string
	var entries []Entry
	for _, entry := range pending {
		entryKeys := stampKeysOf(entry, prepared)
		if len(entryKeys) == 0 {
			continue
		}

		entries = append(entries, entry)
		keys = append(keys, entryKeys...)
	}

	return uniqueKeys(keys), reportsOfEntries(entries), nil
}

func stampKeysOf(entry Entry, prepared PreparedRecord) []string {
	if entry.Finding.Target.Kind() == deltas.TargetBaseline {
		return prepared.Change.PromisesUsing(entry.Runners)
	}
	if entry.Path != prepared.File {
		return nil
	}

	return []string{stampKey(entry.Finding.Target)}
}

func uniqueKeys(keys []string) []string {
	slices.Sort(keys)

	return slices.Compact(keys)
}

func keysOf(asking []deltas.Finding) []string {
	keys := make([]string, 0, len(asking))

	for _, finding := range asking {
		keys = append(keys, stampKey(finding.Target))
	}

	return keys
}

func reportsOfEntries(entries []Entry) []ChangeReport {
	reports := make([]ChangeReport, 0, len(entries))

	for _, entry := range entries {
		reports = append(reports, ReportOf(entry.Finding))
	}

	return reports
}

func (r Recorder) stamps(
	person Person,
	keys []string,
) (map[string]kernel.ApprovalInput, error) {
	stamps := make(map[string]kernel.ApprovalInput, len(keys))

	for _, key := range keys {
		stamp, err := deltas.StampApproval(person.Actor, person.Human, r.Today)
		if err != nil {
			return nil, err
		}

		stamps[key] = stamp
	}

	return stamps, nil
}

func approvalRequired(asking []deltas.Finding) RecordError {
	return recordErrors.Fail(ApprovalRequired, map[string]any{
		"changes": reportsOf(asking),
	})
}

func stampKey(target deltas.ChangeTarget) string {
	if target.Kind() == deltas.TargetConfig {
		return "config"
	}

	return target.Value()
}

func askingFindings(findings []deltas.Finding) []deltas.Finding {
	var asking []deltas.Finding

	for _, finding := range findings {
		if finding.NeedsYes {
			asking = append(asking, finding)
		}
	}

	return asking
}

func resultOf(
	prepared PreparedRecord,
	findings []deltas.Finding,
) RecordResult {
	return RecordResult{
		Type:      prepared.Type,
		ID:        prepared.ID,
		File:      prepared.File,
		Unchanged: prepared.Unchanged,
		Changes:   reportsOf(findings),
		Ignored:   prepared.Ignored,
	}
}

func reportsOf(findings []deltas.Finding) []ChangeReport {
	reports := make([]ChangeReport, 0, len(findings))

	for _, finding := range findings {
		reports = append(reports, ReportOf(finding))
	}

	return reports
}

func ReportOf(finding deltas.Finding) ChangeReport {
	return ChangeReport{
		Target:   string(finding.Target.Kind()),
		ID:       stampKey(finding.Target),
		Level:    int(finding.Level),
		NeedsYes: finding.NeedsYes,
		Reason:   string(finding.Reason),
		Before:   finding.Before,
		After:    finding.After,
	}
}
