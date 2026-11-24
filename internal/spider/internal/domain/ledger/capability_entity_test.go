package ledger_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
)

func sampleCapabilityInput() ledger.CapabilityInput {
	return ledger.CapabilityInput{
		ID:     must(ledger.ParseCapabilityID("bc-7f3a")),
		Author: must(kernel.ParseActor("agent:sample-agent")),
		Name:   must(kernel.ParseRecordText("Login")),
		Intent: must(kernel.ParseRecordText("Users sign in safely")),
		Paths: []kernel.PathGlob{
			must(kernel.ParsePathGlob("src/auth/**")),
		},
		Serves: must(ledger.ParseGoalID("bc-0001")),
		Nongoals: []kernel.RecordText{
			must(kernel.ParseRecordText("No social login")),
		},
	}
}

func samplePromiseInput(id string) ledger.PromiseInput {
	check := must(ledger.ParsePromiseCheck(ledger.PromiseCheckInput{
		Runner: "test-node",
		Params: map[string]string{"target": "auth/lockout"},
	}))

	return ledger.PromiseInput{
		ID:     must(ledger.ParsePromiseID(id)),
		Text:   must(kernel.ParseRecordText("Locks after 5 attempts")),
		Check:  check,
		Origin: must(ledger.ParseOriginWorkID("bc-91cd")),
		Refers: []ledger.PromiseID{
			must(ledger.ParsePromiseID("bc-d4e5")),
		},
	}
}

func sampleCapability(t *testing.T) ledger.Capability {
	t.Helper()

	capability, err := ledger.NewCapability(
		sampleCapabilityInput(),
		sampleNow,
	)
	if err != nil {
		t.Fatal(err)
	}

	return capability
}

func TestNewCapabilityStartsWithoutPromises(t *testing.T) {
	capability := sampleCapability(t)

	if len(capability.Promises()) != 0 {
		t.Fatalf("promises = %v", capability.Promises())
	}
	if capability.Created().String() != "2026-11-22" {
		t.Fatalf("created = %q", capability.Created())
	}
	if serves, found := capability.Serves(); !found {
		t.Fatalf("serves = %q", serves)
	}
}

func TestNewCapabilityRejectsMissingField(t *testing.T) {
	input := sampleCapabilityInput()
	input.Intent = kernel.RecordText{}

	_, err := ledger.NewCapability(input, sampleNow)

	requireCode(t, err, ledger.CapabilityFieldMissing)
}

func TestProposePromiseReturnsNewCapability(t *testing.T) {
	capability := sampleCapability(t)

	next, err := capability.ProposePromise(samplePromiseInput("bc-c3d4"))

	if err != nil {
		t.Fatal(err)
	}
	if len(capability.Promises()) != 0 {
		t.Fatal("original capability changed")
	}

	promises := next.Promises()
	if len(promises) != 1 {
		t.Fatalf("promises = %d", len(promises))
	}
	if promises[0].Status() != ledger.PromiseProposed {
		t.Fatalf("status = %q", promises[0].Status())
	}
	if !next.SameAs(capability) {
		t.Fatal("identity changed")
	}
}

func TestProposePromiseRejectsRepeatedID(t *testing.T) {
	input := samplePromiseInput("bc-c3d4")
	capability := must(sampleCapability(t).ProposePromise(input))

	_, err := capability.ProposePromise(input)

	requireCode(t, err, ledger.PromiseRepeated)
}

func TestProposePromiseRejectsMissingField(t *testing.T) {
	input := samplePromiseInput("bc-c3d4")
	input.Check = ledger.PromiseCheck{}

	_, err := sampleCapability(t).ProposePromise(input)

	requireCode(t, err, ledger.PromiseFieldMissing)
}

func TestProposePromiseDoesNotShareSlices(t *testing.T) {
	first := must(sampleCapability(t).ProposePromise(
		samplePromiseInput("bc-c3d4"),
	))
	left := must(first.ProposePromise(samplePromiseInput("bc-e5f6")))
	right := must(first.ProposePromise(samplePromiseInput("bc-a7b8")))

	if left.Promises()[1].ID().String() != "bc-e5f6" {
		t.Fatalf("left = %q", left.Promises()[1].ID())
	}
	if right.Promises()[1].ID().String() != "bc-a7b8" {
		t.Fatalf("right = %q", right.Promises()[1].ID())
	}
}

func TestPromiseProtectsItsRefers(t *testing.T) {
	input := samplePromiseInput("bc-c3d4")
	capability := must(sampleCapability(t).ProposePromise(input))
	replacement := must(ledger.ParsePromiseID("bc-ffff"))

	input.Refers[0] = replacement
	capability.Promises()[0].Refers()[0] = replacement

	got := capability.Promises()[0].Refers()[0]
	if got.String() != "bc-d4e5" {
		t.Fatalf("refers = %q", got)
	}
}

func TestRestoreCapabilityDoesNotShareState(t *testing.T) {
	refers := []ledger.PromiseID{must(ledger.ParsePromiseID("bc-d4e5"))}
	state := ledger.CapabilityState{
		Promises: []ledger.PromiseState{{Refers: refers}},
	}
	capability := ledger.RestoreCapability(state)

	refers[0] = must(ledger.ParsePromiseID("bc-ffff"))

	got := capability.Promises()[0].Refers()[0]
	if got.String() != "bc-d4e5" {
		t.Fatalf("refers = %q", got)
	}
}

func TestValidateCapabilityStateRejectsAttestedAutomatic(t *testing.T) {
	capability := must(sampleCapability(t).ProposePromise(
		samplePromiseInput("bc-c3d4"),
	))
	attested := must(ledger.ParsePromiseAttestation(
		ledger.PromiseAttestationInput{
			By:     "human:jane",
			At:     "2026-11-22",
			Commit: "9f3c2ab",
		},
	))
	promise := capability.Promises()[0]

	err := ledger.ValidateCapabilityState(ledger.CapabilityState{
		ID:      capability.ID(),
		Created: capability.Created(),
		Author:  capability.Author(),
		Name:    capability.Name(),
		Intent:  capability.Intent(),
		Promises: []ledger.PromiseState{{
			ID:       promise.ID(),
			Text:     promise.Text(),
			Check:    promise.Check(),
			Status:   promise.Status(),
			Attested: attested,
		}},
	})

	requireCode(t, err, ledger.PromiseAttestManual)
}

func TestValidateCapabilityStateAcceptsAttestedManual(t *testing.T) {
	capability := sampleCapability(t)
	manual := must(ledger.ParsePromiseCheck(ledger.PromiseCheckInput{
		Manual: "Open the login page",
	}))
	attested := must(ledger.ParsePromiseAttestation(
		ledger.PromiseAttestationInput{
			By:     "human:jane",
			At:     "2026-11-22",
			Commit: "9f3c2ab",
		},
	))

	err := ledger.ValidateCapabilityState(ledger.CapabilityState{
		ID:      capability.ID(),
		Created: capability.Created(),
		Author:  capability.Author(),
		Name:    capability.Name(),
		Intent:  capability.Intent(),
		Promises: []ledger.PromiseState{{
			ID:       must(ledger.ParsePromiseID("bc-c3d4")),
			Text:     must(kernel.ParseRecordText("Looks right on mobile")),
			Check:    manual,
			Status:   ledger.PromiseActive,
			Attested: attested,
		}},
	})

	if err != nil {
		t.Fatal(err)
	}
}
