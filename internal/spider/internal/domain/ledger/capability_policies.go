package ledger

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
	case isMissing(state.ID):
		return "id"
	case isMissing(state.Created):
		return "created"
	case isMissing(state.Author):
		return "author"
	case isMissing(state.Name):
		return "name"
	case isMissing(state.Intent):
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

	if !isMissing(state.Attested) && !state.Check.IsManual() {
		return invalidCapabilityValue(
			PromiseAttestManual,
			state.ID.String(),
		)
	}

	return nil
}

func missingPromiseField(state PromiseState) string {
	switch {
	case isMissing(state.ID):
		return "id"
	case isMissing(state.Text):
		return "text"
	case state.Check.IsZero():
		return "check"
	case isMissing(state.Status):
		return "status"
	}

	return ""
}
