package action

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type ManualPromise struct {
	ID         string
	Capability string
	Text       string
	Manual     string
	State      string
	Drifted    bool
}

type AttestPorts interface {
	ManualPromises(
		ctx context.Context,
		revision string,
	) ([]ManualPromise, error)
	Attest(
		ctx context.Context,
		promise string,
		stamp kernel.ApprovalInput,
		commit string,
		evidence string,
	) error
	Head(ctx context.Context) (string, error)
}

type AttestInput struct {
	Promises []string
	Evidence string
	Base     string
}

type AttestResult struct {
	Pending  []ManualPromise
	Attested []string
	Commit   string
}

type Attester struct {
	Ports    AttestPorts
	Base     BaseResolver
	Identity Identity
	Today    kernel.CalendarDate
}

func (a Attester) Attest(
	ctx context.Context,
	input AttestInput,
) (AttestResult, error) {
	if len(input.Promises) == 0 {
		return a.pending(ctx, input.Base)
	}

	person, err := a.Identity.Whoami(ctx)
	if err != nil {
		return AttestResult{}, err
	}

	stamp, err := deltas.StampApproval(person.Actor, person.Human, a.Today)
	if err != nil {
		return AttestResult{}, err
	}

	head, err := a.Ports.Head(ctx)
	if err != nil {
		return AttestResult{}, err
	}

	result := AttestResult{Attested: []string{}, Commit: head}

	for _, promise := range input.Promises {
		err := a.Ports.Attest(ctx, promise, stamp, head, input.Evidence)
		if err != nil {
			return result, err
		}

		result.Attested = append(result.Attested, promise)
	}

	return result, nil
}

func (a Attester) pending(
	ctx context.Context,
	forced string,
) (AttestResult, error) {
	base, err := a.Base.Base(ctx, forced)
	if err != nil {
		return AttestResult{}, err
	}

	promises, err := a.Ports.ManualPromises(ctx, base)

	return AttestResult{Pending: promises, Attested: []string{}}, err
}
