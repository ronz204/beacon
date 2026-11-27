package weaver

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/action"
)

func (w Weaver) Close(
	ctx context.Context,
	id string,
	reason string,
	approve bool,
) (CloseView, error) {
	result, err := w.closer.Close(ctx, action.CloseInput{
		ID:      id,
		Reason:  reason,
		Approve: approve,
	})

	return CloseView{
		Type:    result.Type,
		ID:      result.ID,
		File:    result.File,
		Closed:  result.Closed,
		Changes: changeViews(result.Changes),
	}, err
}
