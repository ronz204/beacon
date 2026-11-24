package action

import (
	"context"
	"slices"

	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/thread"
)

type Verifications map[ledger.PromiseID]ledger.Verification

func Verify(
	ctx context.Context,
	snapshot Snapshot,
	proofs ledger.ProofReader,
) (Verifications, []error) {
	var failures []error

	outcomes, err := proofs.ReadOutcomes(ctx)
	if err != nil {
		failures = append(failures, err)
	}

	verifications := make(Verifications)
	for _, capability := range snapshot.Capabilities {
		for _, promise := range capability.Promises() {
			outcome, found := outcomes[promise.ID()]
			verifications[promise.ID()] = ledger.VerifyPromise(
				promise,
				outcome,
				found,
			)
		}
	}

	return verifications, failures
}

func Signals(
	snapshot Snapshot,
	graph thread.Graph,
	verifications Verifications,
) []thread.Signal {
	signals := thread.BlockedSignals(graph)

	for _, capability := range snapshot.Capabilities {
		scoped := len(capability.Paths()) > 0

		for _, promise := range capability.Promises() {
			verification := verifications[promise.ID()]
			found := promiseSignals(promise, scoped, verification)
			signals = append(signals, found...)
		}
	}

	slices.SortFunc(signals, thread.CompareSignals)

	return signals
}

func promiseSignals(
	promise ledger.Promise,
	scoped bool,
	verification ledger.Verification,
) []thread.Signal {
	id, err := recordIDOf(promise.ID())
	if err != nil {
		return nil
	}

	var kinds []thread.SignalKind
	if verification.Status() == ledger.VerificationViolated {
		kinds = append(kinds, thread.SignalRed)
	}
	if verification.NotRunHere() {
		kinds = append(kinds, thread.SignalNotRunHere)
	}
	if !scoped {
		kinds = append(kinds, thread.SignalUnscoped)
	}

	signals := make([]thread.Signal, 0, len(kinds))
	for _, kind := range kinds {
		signals = append(signals, thread.Signal{Kind: kind, Subject: id})
	}

	return signals
}
