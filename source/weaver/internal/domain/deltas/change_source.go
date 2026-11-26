package deltas

import "context"

type ChangeSource interface {
	ListChanges(ctx context.Context, revision string) ([]StateChange, error)
}
