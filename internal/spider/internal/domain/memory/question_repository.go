package memory

import "context"

type QuestionRepository interface {
	ListQuestions(ctx context.Context) ([]Question, []error)
}
