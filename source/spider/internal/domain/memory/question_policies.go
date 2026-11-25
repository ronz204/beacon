package memory

import "github.com/ronz204/beacon/source/kernel"

func ValidateQuestionState(state QuestionState) error {
	field := missingQuestionField(state)
	if field != "" {
		details := map[string]any{"field": field}
		return questionErrors.Fail(QuestionFieldMissing, details)
	}

	if state.Status == QuestionAnswered && kernel.IsMissing(state.Answer) {
		return invalidQuestion(QuestionAnswerMissing, state.ID.String())
	}

	return nil
}

func missingQuestionField(state QuestionState) string {
	switch {
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Created):
		return "created"
	case kernel.IsMissing(state.Author):
		return "author"
	case kernel.IsMissing(state.Text):
		return "text"
	case kernel.IsMissing(state.Status):
		return "status"
	}

	return ""
}
