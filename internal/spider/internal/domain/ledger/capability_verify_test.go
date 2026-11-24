package ledger_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
)

func manualPromise(t *testing.T, attested bool) ledger.Promise {
	t.Helper()

	check := must(ledger.ParsePromiseCheck(ledger.PromiseCheckInput{
		Manual: "Open the login page",
	}))
	state := ledger.PromiseState{
		ID:     must(ledger.ParsePromiseID("bc-c4a1")),
		Text:   must(kernel.ParseRecordText("Looks right")),
		Check:  check,
		Status: ledger.PromiseActive,
	}

	if attested {
		state.Attested = must(ledger.ParsePromiseAttestation(
			ledger.PromiseAttestationInput{
				By:     "human:jane",
				At:     "2026-11-22",
				Commit: "9f3c2ab",
			},
		))
	}

	capability := ledger.RestoreCapability(ledger.CapabilityState{
		Promises: []ledger.PromiseState{state},
	})

	return capability.Promises()[0]
}

func automaticPromise(t *testing.T) ledger.Promise {
	t.Helper()

	capability := must(sampleCapability(t).ProposePromise(
		samplePromiseInput("bc-c3d4"),
	))

	return capability.Promises()[0]
}

func TestVerifyPromise(t *testing.T) {
	cases := []struct {
		name       string
		promise    ledger.Promise
		outcome    ledger.ProofOutcome
		found      bool
		status     ledger.VerificationStatus
		notRunHere bool
	}{
		{
			"passed check",
			automaticPromise(t),
			ledger.ProofPassed,
			true,
			ledger.VerificationHolds,
			false,
		},
		{
			"failed check",
			automaticPromise(t),
			ledger.ProofFailed,
			true,
			ledger.VerificationViolated,
			false,
		},
		{
			"no result here",
			automaticPromise(t),
			"",
			false,
			ledger.VerificationUnverified,
			true,
		},
		{
			"manual attested",
			manualPromise(t, true),
			"",
			false,
			ledger.VerificationAttested,
			false,
		},
		{
			"manual pending",
			manualPromise(t, false),
			"",
			false,
			ledger.VerificationUnverified,
			false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := ledger.ProofResult{Outcome: c.outcome, Commit: "9f3c2ab"}
			got := ledger.VerifyPromise(c.promise, result, c.found)

			if got.Status() != c.status {
				t.Fatalf("status = %q, want %q", got.Status(), c.status)
			}
			if got.NotRunHere() != c.notRunHere {
				t.Fatalf("not run here = %v", got.NotRunHere())
			}
			if got.Drifted() {
				t.Fatal("verification must start without drift")
			}
		})
	}
}

func TestVerificationKeepsEvidenceCommit(t *testing.T) {
	passed := ledger.ProofResult{Outcome: ledger.ProofPassed, Commit: "abc1234"}

	holds := ledger.VerifyPromise(automaticPromise(t), passed, true)
	attested := ledger.VerifyPromise(manualPromise(t, true), passed, false)
	violated := ledger.VerifyPromise(
		automaticPromise(t),
		ledger.ProofResult{Outcome: ledger.ProofFailed, Commit: "abc1234"},
		true,
	)

	if holds.Commit() != "abc1234" {
		t.Fatalf("holds commit = %q", holds.Commit())
	}
	if attested.Commit() != "9f3c2ab" {
		t.Fatalf("attested commit = %q", attested.Commit())
	}
	if violated.Commit() != "" {
		t.Fatalf("violated commit = %q", violated.Commit())
	}
	if !holds.WithDrift().Drifted() {
		t.Fatal("WithDrift must mark drift")
	}
}
