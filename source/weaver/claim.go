package weaver

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/action"
)

func (w Weaver) Claim(
	ctx context.Context,
	task string,
	force bool,
) (ClaimView, error) {
	result, err := w.claimer.Claim(ctx, action.ClaimInput{
		Task:  task,
		Force: force,
	})

	view := ClaimView{
		Task:     result.Task,
		Branch:   result.Branch,
		Worktree: result.Worktree,
		Owner:    result.Owner,
	}
	if result.Replaced != nil {
		holder := HolderView(*result.Replaced)
		view.Replaced = &holder
	}

	return view, err
}
