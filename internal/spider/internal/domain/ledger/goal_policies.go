package ledger

func ValidateGoalState(state GoalState) error {
	field := missingGoalField(state)

	if field != "" {
		details := map[string]any{"field": field}
		return goalErrors.Fail(GoalFieldMissing, details)
	}

	return nil
}

func missingGoalField(state GoalState) string {
	switch {
	case isMissing(state.ID):
		return "id"
	case isMissing(state.Created):
		return "created"
	case isMissing(state.Author):
		return "author"
	case isMissing(state.Statement):
		return "statement"
	case isMissing(state.Success):
		return "success"
	case isMissing(state.Status):
		return "status"
	}

	return ""
}

func isMissing[T comparable](value T) bool {
	var zero T
	return value == zero
}
