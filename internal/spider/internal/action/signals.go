package action

import (
	"context"

	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/thread"
)

type Verifications map[ledger.PromiseID]ledger.Verification

func verify(
	ctx context.Context,
	snapshot Snapshot,
	proofs ledger.ProofReader,
) (Verifications, []error) {
	var failures []error

	results, err := proofs.ReadResults(ctx)
	if err != nil {
		failures = append(failures, err)
	}

	verifications := make(Verifications)
	for _, capability := range snapshot.Capabilities {
		for _, promise := range capability.Promises() {
			result, found := results[promise.ID()]
			verifications[promise.ID()] = ledger.VerifyPromise(
				promise,
				result,
				found,
			)
		}
	}

	return verifications, failures
}

func promiseSignalsOf(
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

	var signals []thread.Signal
	add := func(kind thread.SignalKind, applies bool) {
		if applies {
			signals = append(signals, thread.Signal{Kind: kind, Subject: id})
		}
	}

	add(thread.SignalRed, verification.Status() == ledger.VerificationViolated)
	add(thread.SignalNotRunHere, verification.NotRunHere())
	add(thread.SignalDrift, verification.Drifted())
	add(thread.SignalUnscoped, !scoped)

	return signals
}
