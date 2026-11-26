package spider

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"

	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

type ClaimView struct {
	Task     string `json:"task"`
	Branch   string `json:"branch"`
	Worktree string `json:"worktree"`
	Day      string `json:"day"`
}

func (p Project) Worktree() string {
	return p.root
}

func (p Project) Claims(ctx context.Context) ([]ClaimView, error) {
	local, err := states.Open(ctx, p.root)
	if err != nil {
		return nil, err
	}

	claims, err := local.ListClaims(ctx)
	if err != nil {
		return nil, err
	}

	views := make([]ClaimView, 0, len(claims))
	for _, claim := range claims {
		views = append(views, ClaimView{
			Task:     claim.Task.String(),
			Branch:   claim.Branch.String(),
			Worktree: claim.Worktree,
			Day:      claim.Day.String(),
		})
	}

	return views, nil
}

func (p Project) SaveClaim(ctx context.Context, view ClaimView) error {
	task, err := agenda.ParseTaskID(view.Task)
	if err != nil {
		return err
	}

	branch, err := agenda.ParseTaskBranch(view.Branch)
	if err != nil {
		return err
	}

	day := today()
	if view.Day != "" {
		day, err = kernel.ParseCalendarDate(view.Day)
		if err != nil {
			return err
		}
	}

	local, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	return local.SaveClaim(ctx, agenda.TaskClaim{
		Task:     task,
		Branch:   branch,
		Worktree: view.Worktree,
		Day:      day,
	})
}

func (p Project) RemoveClaim(ctx context.Context, taskID string) error {
	task, err := agenda.ParseTaskID(taskID)
	if err != nil {
		return err
	}

	local, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	return local.RemoveClaims(ctx, []agenda.TaskID{task})
}
