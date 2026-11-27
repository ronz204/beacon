package weaver

import (
	"context"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

func (w Weaver) Brief(ctx context.Context) (Summary, error) {
	return w.summary(ctx, w.project.Brief)
}

func (w Weaver) Status(ctx context.Context) (Summary, error) {
	return w.summary(ctx, w.project.Status)
}

func (w Weaver) summary(
	ctx context.Context,
	build func(context.Context, []string) (spider.BriefView, []spider.Finding),
) (Summary, error) {
	pending, err := w.differ.Pending(ctx)
	checked := !action.IsBaseMissing(err)
	if err != nil && checked {
		return Summary{}, err
	}

	lines := make([]string, 0, len(pending))
	for _, entry := range pending {
		lines = append(lines, action.LineOf(entry))
	}

	view, findings := build(ctx, lines)
	view.Signals = append(view.Signals, pendingSignals(pending)...)

	return Summary{View: view, Findings: findings, PendingChecked: checked}, nil
}
