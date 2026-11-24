package agenda

import "context"

type WorkspaceReader interface {
	ListBranches(ctx context.Context) ([]TaskBranch, error)
	WorktreeExists(ctx context.Context, path string) (bool, error)
}
