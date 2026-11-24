package agenda

import "github.com/ronz204/beacon/internal/kernel"

func ValidateWorkState(state WorkState) error {
	field := missingWorkField(state)

	if field != "" {
		details := map[string]any{"field": field}
		return workErrors.Fail(WorkFieldMissing, details)
	}

	return nil
}

func missingWorkField(state WorkState) string {
	switch {
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Created):
		return "created"
	case kernel.IsMissing(state.Author):
		return "author"
	case kernel.IsMissing(state.Request):
		return "request"
	case kernel.IsMissing(state.Status):
		return "status"
	}

	return ""
}
