package action

import (
	"context"
	"path/filepath"
	"slices"

	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
)

func PruneClaims(
	ctx context.Context,
	claims agenda.ClaimRepository,
	workspace agenda.WorkspaceReader,
) ([]agenda.TaskClaim, error) {
	current, err := claims.ListClaims(ctx)
	if err != nil {
		return nil, err
	}

	branches, err := workspace.ListBranches(ctx)
	if err != nil {
		return nil, err
	}

	worktrees, err := workspace.ListWorktrees(ctx)
	if err != nil {
		return nil, err
	}

	var removed []agenda.TaskClaim
	var tasks []agenda.TaskID
	for _, claim := range current {
		if isClaimAlive(claim, branches, worktrees) {
			continue
		}

		removed = append(removed, claim)
		tasks = append(tasks, claim.Task)
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	return removed, claims.RemoveClaims(ctx, tasks)
}

func isClaimAlive(
	claim agenda.TaskClaim,
	branches []agenda.TaskBranch,
	worktrees []string,
) bool {
	if !slices.Contains(branches, claim.Branch) {
		return false
	}

	worktree := filepath.Clean(claim.Worktree)

	return slices.ContainsFunc(worktrees, func(path string) bool {
		return samePath(filepath.Clean(path), worktree)
	})
}

func samePath(a, b string) bool {
	return filepath.ToSlash(a) == filepath.ToSlash(b)
}
