package facade

import (
	"context"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

type Tasks struct {
	project spider.Project
}

func NewTasks(project spider.Project) Tasks {
	return Tasks{project: project}
}

func (t Tasks) Task(
	ctx context.Context,
	id string,
) (action.StoredTask, error) {
	stored, err := t.project.Get(ctx, id)
	if err != nil {
		return action.StoredTask{}, err
	}

	return action.StoredTask{
		Type:     stored.Type,
		Data:     stored.Data,
		Archived: stored.Archived,
	}, nil
}

func (t Tasks) Branch(ctx context.Context) (action.BranchState, error) {
	branch, err := t.project.Branch(ctx)

	return action.BranchState{Name: branch.Name, Detached: branch.Detached}, err
}

func (t Tasks) Worktree() string {
	return t.project.Worktree()
}

func (t Tasks) Claims(ctx context.Context) ([]action.LocalClaim, error) {
	views, err := t.project.Claims(ctx)
	if err != nil {
		return nil, err
	}

	claims := make([]action.LocalClaim, 0, len(views))
	for _, view := range views {
		claims = append(claims, action.LocalClaim(view))
	}

	return claims, nil
}

func (t Tasks) SaveClaim(ctx context.Context, claim action.LocalClaim) error {
	return t.project.SaveClaim(ctx, spider.ClaimView(claim))
}

func (t Tasks) RemoveClaim(ctx context.Context, task string) error {
	return t.project.RemoveClaim(ctx, task)
}
