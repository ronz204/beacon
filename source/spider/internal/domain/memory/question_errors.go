package memory

import "github.com/ronz204/beacon/source/kernel"

type QuestionCode string

const (
	QuestionIDInvalid     QuestionCode = "question_id_invalid"
	QuestionStatusInvalid QuestionCode = "question_status_invalid"
	QuestionBlockInvalid  QuestionCode = "question_block_invalid"
	QuestionOriginInvalid QuestionCode = "question_origin_invalid"
	QuestionFieldMissing  QuestionCode = "question_field_missing"
	QuestionAnswerMissing QuestionCode = "question_answer_missing"
)

type QuestionError = kernel.Error[QuestionCode]

var questionErrors = kernel.ErrorCatalog[QuestionCode]{
	QuestionIDInvalid:     "question id must be a root record id",
	QuestionStatusInvalid: "question status must be open or answered",
	QuestionBlockInvalid:  "question blocks must be record ids",
	QuestionOriginInvalid: "question origin must be a root work id",
	QuestionFieldMissing:  "question is missing a required field",
	QuestionAnswerMissing: "an answered question needs an answer",
}

func invalidQuestion(code QuestionCode, raw string) QuestionError {
	details := map[string]any{"value": raw}
	return questionErrors.Fail(code, details)
}
