package weaver

import (
	"context"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
	"github.com/ronz204/beacon/source/weaver/internal/driver/facade"
	"github.com/ronz204/beacon/source/weaver/internal/driver/identity"
)

type ChangeView struct {
	Target   string `json:"target"`
	ID       string `json:"id"`
	Level    int    `json:"level"`
	NeedsYes bool   `json:"needsYes"`
	Reason   string `json:"reason,omitempty"`
	Before   string `json:"before,omitempty"`
	After    string `json:"after,omitempty"`
}

type RecordView struct {
	Type      string       `json:"type"`
	ID        string       `json:"id"`
	File      string       `json:"file"`
	Saved     bool         `json:"saved"`
	Unchanged bool         `json:"unchanged"`
	Changes   []ChangeView `json:"changes"`
	Ignored   []string     `json:"ignored"`
}

type Weaver struct {
	changes deltas.ChangeSource
	records action.Recorder
}

func New(project spider.Project) Weaver {
	today, _ := kernel.CalendarDateOf(time.Now())

	return Weaver{
		changes: facade.NewChanges(project),
		records: action.Recorder{
			Store:    facade.NewRecords(project),
			Identity: identity.Git{},
			Today:    today,
		},
	}
}

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
		Changes:   make([]ChangeView, 0, len(result.Changes)),
		Ignored:   result.Ignored,
	}
	if view.Ignored == nil {
		view.Ignored = []string{}
	}

	for _, change := range result.Changes {
		view.Changes = append(view.Changes, ChangeView(change))
	}

	return view
}
