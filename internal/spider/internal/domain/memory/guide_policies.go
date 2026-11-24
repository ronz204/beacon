package memory

import "github.com/ronz204/beacon/internal/kernel"

func ValidateGuideState(state GuideState) error {
	field := missingGuideField(state)

	if field != "" {
		details := map[string]any{"field": field}
		return guideErrors.Fail(GuideFieldMissing, details)
	}

	return nil
}

func missingGuideField(state GuideState) string {
	switch {
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Title):
		return "title"
	case kernel.IsMissing(state.Kind):
		return "kind"
	}

	return ""
}
