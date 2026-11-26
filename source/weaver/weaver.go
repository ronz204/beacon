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

type CloseView struct {
	Type    string       `json:"type"`
	ID      string       `json:"id"`
	File    string       `json:"file"`
	Closed  bool         `json:"closed"`
	Changes []ChangeView `json:"changes"`
}

type HolderView struct {
	Source   string `json:"source"`
	Branch   string `json:"branch"`
	Worktree string `json:"worktree,omitempty"`
	Owner    string `json:"owner,omitempty"`
}

type ClaimView struct {
	Task     string      `json:"task"`
	Branch   string      `json:"branch"`
	Worktree string      `json:"worktree"`
	Owner    string      `json:"owner"`
	Replaced *HolderView `json:"replaced,omitempty"`
}

type Weaver struct {
	changes deltas.ChangeSource
	records action.Recorder
	closer  action.Closer
	claimer action.Claimer
}

func New(project spider.Project) Weaver {
	today, _ := kernel.CalendarDateOf(time.Now())
	records := facade.NewRecords(project)
	who := identity.Git{}

	return Weaver{
		changes: facade.NewChanges(project),
		records: action.Recorder{
			Store:    records,
			Identity: who,
			Today:    today,
		},
		closer: action.Closer{
			Store:    facade.NewCloses(project),
			Identity: who,
			Today:    today,
		},
		claimer: action.Claimer{
			Ports:    facade.NewTasks(project),
			Store:    records,
			Identity: who,
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

func recordView(result action.RecordResult) RecordView {
	view := RecordView{
		Type:      result.Type,
		ID:        result.ID,
		File:      result.File,
		Saved:     result.Saved,
		Unchanged: result.Unchanged,
		Changes:   changeViews(result.Changes),
		Ignored:   result.Ignored,
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
