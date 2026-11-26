package facade

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

type Closes struct {
	project spider.Project
}

func NewCloses(project spider.Project) Closes {
	return Closes{project: project}
}

func (c Closes) PrepareClose(
	ctx context.Context,
	id string,
	reason string,
) (action.PreparedClose, error) {
	prepared, err := c.project.PrepareClose(ctx, id, reason)
	if err != nil {
		return action.PreparedClose{}, err
	}

	kind := "modified"
	if prepared.After == nil {
		kind = "deleted"
	}

	change, err := convertChange(spider.StateChange{
		Path:   prepared.File,
		Kind:   kind,
		Before: prepared.Before,
		After:  prepared.After,
	})
	if err != nil {
		return action.PreparedClose{}, err
	}

	return action.PreparedClose{
		Type:   prepared.Type,
		ID:     prepared.ID,
		File:   prepared.File,
		Change: change,
		Token:  prepared,
	}, nil
}

func (c Closes) Close(
	ctx context.Context,
	prepared action.PreparedClose,
	stamps map[string]kernel.ApprovalInput,
) error {
	token, _ := prepared.Token.(spider.ClosePrepared)

	return c.project.Close(ctx, token, stamps)
}
