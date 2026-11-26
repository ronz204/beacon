package facade

import (
	"context"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

type Records struct {
	project spider.Project
}

func NewRecords(project spider.Project) Records {
	return Records{project: project}
}

func (r Records) Prepare(
	ctx context.Context,
	payload []byte,
	author string,
) (action.PreparedRecord, error) {
	prepared, err := r.project.Prepare(ctx, payload, author)
	if err != nil {
		return action.PreparedRecord{}, err
	}

	kind := "modified"
	if prepared.Before == nil {
		kind = "added"
	}

	change, err := convertChange(spider.StateChange{
		Path:   prepared.File,
		Kind:   kind,
		Before: prepared.Before,
		After:  &prepared.After,
	})
	if err != nil {
		return action.PreparedRecord{}, err
	}

	return action.PreparedRecord{
		Type:      prepared.Type,
		ID:        prepared.ID,
		File:      prepared.File,
		Change:    change,
		Ignored:   prepared.Ignored,
		Unchanged: prepared.Unchanged,
		Token:     prepared,
	}, nil
}

func (r Records) Save(
	ctx context.Context,
	prepared action.PreparedRecord,
	stamps map[string]kernel.ApprovalInput,
) error {
	token, _ := prepared.Token.(spider.Prepared)

	return r.project.Save(ctx, token, stamps)
}
