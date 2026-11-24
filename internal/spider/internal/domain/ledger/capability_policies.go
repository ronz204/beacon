package ledger

import "github.com/ronz204/beacon/internal/kernel"

func ValidateCapabilityState(state CapabilityState) error {
	field := missingCapabilityField(state)
	if field != "" {
		details := map[string]any{"field": field}
		return capabilityErrors.Fail(CapabilityFieldMissing, details)
	}

	seen := make(map[PromiseID]bool, len(state.Promises))
	for _, promise := range state.Promises {
		if err := validatePromiseState(promise); err != nil {
			return err
		}
		if seen[promise.ID] {
			return invalidCapabilityValue(
				PromiseRepeated,
				promise.ID.String(),
			)
		}

		seen[promise.ID] = true
	}

	return nil
}

func missingCapabilityField(state CapabilityState) string {
	switch {
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Created):
		return "created"
	case kernel.IsMissing(state.Author):
		return "author"
	case kernel.IsMissing(state.Name):
		return "name"
	case kernel.IsMissing(state.Intent):
		return "intent"
	}

	return ""
}

func validatePromiseState(state PromiseState) error {
	field := missingPromiseField(state)
	if field != "" {
		details := map[string]any{"field": field}
		return capabilityErrors.Fail(PromiseFieldMissing, details)
	}

	if !kernel.IsMissing(state.Attested) && !state.Check.IsManual() {
		return invalidCapabilityValue(
			PromiseAttestManual,
			state.ID.String(),
		)
	}

	return nil
}

func missingPromiseField(state PromiseState) string {
	switch {
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Text):
		return "text"
	case state.Check.IsZero():
		return "check"
	case kernel.IsMissing(state.Status):
		return "status"
	}

	return ""
}

func VerifyPromise(
	promise Promise,
	outcome ProofOutcome,
	found bool,
) Verification {
	switch {
	case promise.Check().IsManual():
		return manualVerification(promise)
	case !found:
		return Verification{
			status:     VerificationUnverified,
			notRunHere: true,
		}
	case outcome == ProofPassed:
		return Verification{status: VerificationHolds}
	}

	return Verification{status: VerificationViolated}
}

func manualVerification(promise Promise) Verification {
	if _, attested := promise.Attested(); attested {
		return Verification{status: VerificationAttested}
	}

	return Verification{status: VerificationUnverified}
}
