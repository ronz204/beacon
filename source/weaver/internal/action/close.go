package action

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type PreparedClose struct {
	Type   string
	ID     string
	File   string
	Change deltas.StateChange
	Token  any
}

type CloseStore interface {
	PrepareClose(
		ctx context.Context,
		id string,
		reason string,
	) (PreparedClose, error)
	Close(
		ctx context.Context,
		prepared PreparedClose,
		stamps map[string]kernel.ApprovalInput,
	) error
}

type CloseInput struct {
	ID      string
	Reason  string
	Approve bool
}

type CloseResult struct {
	Type    string
	ID      string
	File    string
	Closed  bool
	Changes []ChangeReport
}

type Closer struct {
	Store    CloseStore
	Identity Identity
	Today    kernel.CalendarDate
}

func (c Closer) Close(
	ctx context.Context,
	input CloseInput,
) (CloseResult, error) {
	person, err := c.Identity.Whoami(ctx)
	if err != nil {
		return CloseResult{}, err
	}

	prepared, err := c.Store.PrepareClose(ctx, input.ID, input.Reason)
	if err != nil {
		return CloseResult{}, err
	}

	findings := deltas.ClassifyElements(prepared.Change)
	result := CloseResult{
		Type:    prepared.Type,
		ID:      prepared.ID,
		File:    prepared.File,
		Changes: reportsOf(findings),
	}

	asking := askingFindings(findings)
	if len(asking) > 0 && !input.Approve {
		return result, approvalRequired(asking)
	}

	stamps, err := c.stamps(person, asking, prepared.ID)
	if err != nil {
		return result, err
	}

	if err := c.Store.Close(ctx, prepared, stamps); err != nil {
		return result, err
	}

	result.Closed = true

	return result, nil
}

func (c Closer) stamps(
	person Person,
	asking []deltas.Finding,
	id string,
) (map[string]kernel.ApprovalInput, error) {
	if len(asking) == 0 {
		return nil, nil
	}

	stamp, err := deltas.StampApproval(person.Actor, person.Human, c.Today)
	if err != nil {
		return nil, err
	}

	return map[string]kernel.ApprovalInput{id: stamp}, nil
}
