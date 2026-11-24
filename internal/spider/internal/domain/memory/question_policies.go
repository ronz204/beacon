package memory

func ValidateQuestionState(state QuestionState) error {
	field := missingQuestionField(state)
	if field != "" {
		details := map[string]any{"field": field}
		return questionErrors.Fail(QuestionFieldMissing, details)
	}

	if state.Status == QuestionAnswered && isMissing(state.Answer) {
		return invalidQuestion(QuestionAnswerMissing, state.ID.String())
	}

	return nil
}

func missingQuestionField(state QuestionState) string {
	switch {
	case isMissing(state.ID):
		return "id"
	case isMissing(state.Created):
		return "created"
	case isMissing(state.Author):
		return "author"
	case isMissing(state.Text):
		return "text"
	case isMissing(state.Status):
		return "status"
	}

	return ""
}
