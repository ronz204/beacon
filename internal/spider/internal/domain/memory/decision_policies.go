package memory

func ValidateDecisionState(state DecisionState) error {
	field := missingDecisionField(state)

	if field != "" {
		details := map[string]any{"field": field}
		return decisionErrors.Fail(DecisionFieldMissing, details)
	}

	return nil
}

func missingDecisionField(state DecisionState) string {
	switch {
	case isMissing(state.ID):
		return "id"
	case isMissing(state.Created):
		return "created"
	case isMissing(state.Author):
		return "author"
	case isMissing(state.What):
		return "what"
	case isMissing(state.Why):
		return "why"
	case isMissing(state.Status):
		return "status"
	}

	return ""
}

func isMissing[T comparable](value T) bool {
	var zero T
	return value == zero
}
