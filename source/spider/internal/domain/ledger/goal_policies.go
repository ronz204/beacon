package ledger

import "github.com/ronz204/beacon/source/kernel"

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
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Created):
		return "created"
	case kernel.IsMissing(state.Author):
		return "author"
	case kernel.IsMissing(state.Statement):
		return "statement"
	case kernel.IsMissing(state.Success):
		return "success"
	case kernel.IsMissing(state.Status):
		return "status"
	}

	return ""
}
