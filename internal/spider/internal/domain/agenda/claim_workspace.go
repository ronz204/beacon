package agenda

import "context"

type WorkspaceReader interface {
	ListBranches(ctx context.Context) ([]TaskBranch, error)
	ListWorktrees(ctx context.Context) ([]string, error)
}
