package ledger

import "context"

type GoalRepository interface {
	ListGoals(ctx context.Context) ([]Goal, []error)
}
