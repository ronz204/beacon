package memory

import "context"

type DecisionRepository interface {
	ListDecisions(ctx context.Context) ([]Decision, []error)
}
