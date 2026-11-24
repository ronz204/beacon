package memory

import "github.com/ronz204/beacon/internal/kernel"

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
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Created):
		return "created"
	case kernel.IsMissing(state.Author):
		return "author"
	case kernel.IsMissing(state.What):
		return "what"
	case kernel.IsMissing(state.Why):
		return "why"
	case kernel.IsMissing(state.Status):
		return "status"
	}

	return ""
}
