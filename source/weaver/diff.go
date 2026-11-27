package weaver

import (
	"context"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

func (w Weaver) Diff(
	ctx context.Context,
	since string,
	base string,
) (DiffView, []spider.Signal, error) {
	result, err := w.differ.Diff(ctx, action.DiffInput{
		Since: since,
		Base:  base,
	})

	return diffView(result), pendingSignals(result.Pending), err
}

func diffView(result action.DiffResult) DiffView {
	view := DiffView{
		Revision: result.Revision,
		Levels:   []LevelView{},
		Pending:  entryViews(result.Pending),
	}

	for _, entry := range entryViews(result.Entries) {
		last := len(view.Levels) - 1
		if last < 0 || view.Levels[last].Level != entry.Level {
			view.Levels = append(view.Levels, LevelView{Level: entry.Level})
			last++
		}

		view.Levels[last].Changes = append(view.Levels[last].Changes, entry)
	}

	return view
}

func entryViews(entries []action.Entry) []EntryView {
	views := make([]EntryView, 0, len(entries))

	for _, entry := range entries {
		report := action.ReportOf(entry.Finding)
		views = append(views, EntryView{
			Path:     entry.Path,
			Kind:     entry.Kind,
			Target:   report.Target,
			ID:       report.ID,
			Level:    report.Level,
			Reason:   report.Reason,
			Before:   report.Before,
			After:    report.After,
			NeedsYes: report.NeedsYes,
			Approved: entry.Approved,
		})
	}

	return views
}

const pendingSignal = "pending_yes"

func pendingSignals(pending []action.Entry) []spider.Signal {
	signals := make([]spider.Signal, 0, len(pending))

	for _, entry := range pending {
		signals = append(signals, spider.Signal{
			Kind:   pendingSignal,
			Record: action.ReportOf(entry.Finding).ID,
		})
	}

	return signals
}
