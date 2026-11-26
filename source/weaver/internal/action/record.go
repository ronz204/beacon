package action

import (
	"context"

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

type Recorder struct {
	Store    RecordStore
	Identity Identity
	Today    kernel.CalendarDate
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

	if prepared.Unchanged {
		return result, nil
	}

	asking := askingFindings(findings)
	if len(asking) > 0 && !input.Approve {
		return result, recordErrors.Fail(ApprovalRequired, map[string]any{
			"changes": reportsOf(asking),
		})
	}

	stamps, err := r.stamps(person, asking)
	if err != nil {
		return result, err
	}

	if err := r.Store.Save(ctx, prepared, stamps); err != nil {
		return result, err
	}

	result.Saved = true

	return result, nil
}

func (r Recorder) stamps(
	person Person,
	asking []deltas.Finding,
) (map[string]kernel.ApprovalInput, error) {
	stamps := make(map[string]kernel.ApprovalInput, len(asking))

	for _, finding := range asking {
		stamp, err := deltas.StampApproval(person.Actor, person.Human, r.Today)
		if err != nil {
			return nil, err
		}

		stamps[stampKey(finding.Target)] = stamp
	}

	return stamps, nil
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
		reports = append(reports, ChangeReport{
			Target:   string(finding.Target.Kind()),
			ID:       stampKey(finding.Target),
			Level:    int(finding.Level),
			NeedsYes: finding.NeedsYes,
		})
	}

	return reports
}
