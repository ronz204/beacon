package memory

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
	case isMissing(state.ID):
		return "id"
	case isMissing(state.Title):
		return "title"
	case isMissing(state.Kind):
		return "kind"
	}

	return ""
}
