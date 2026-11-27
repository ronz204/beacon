package weaver

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/action"
)

func (w Weaver) Record(
	ctx context.Context,
	payload []byte,
	approve bool,
) (RecordView, error) {
	result, err := w.records.Record(ctx, action.RecordInput{
		Payload: payload,
		Approve: approve,
	})

	return recordView(result), err
}

func recordView(result action.RecordResult) RecordView {
	view := RecordView{
		Type:      result.Type,
		ID:        result.ID,
		File:      result.File,
		Saved:     result.Saved,
		Unchanged: result.Unchanged,
		Changes:   changeViews(result.Changes),
		Ignored:   result.Ignored,
		Guidance:  result.Guidance,
	}
	if view.Ignored == nil {
		view.Ignored = []string{}
	}

	return view
}

func changeViews(changes []action.ChangeReport) []ChangeView {
	views := make([]ChangeView, 0, len(changes))

	for _, change := range changes {
		views = append(views, ChangeView(change))
	}

	return views
}
