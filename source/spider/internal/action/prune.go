package action

import (
	"context"
	"slices"

	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
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

	var removed []agenda.TaskClaim
	var tasks []agenda.TaskID
	for _, claim := range current {
		alive, err := isClaimAlive(ctx, claim, branches, workspace)
		if err != nil {
			return nil, err
		}
		if alive {
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
	ctx context.Context,
	claim agenda.TaskClaim,
	branches []agenda.TaskBranch,
	workspace agenda.WorkspaceReader,
) (bool, error) {
	if !slices.Contains(branches, claim.Branch) {
		return false, nil
	}

	return workspace.WorktreeExists(ctx, claim.Worktree)
}
