package weaver

import (
	"context"
	"io"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/driver/runner"
)

func (w Weaver) Streaming(out io.Writer) Weaver {
	w.checker.Executor = runner.New(out)

	return w
}

func (w Weaver) Check(
	ctx context.Context,
	capability string,
	all bool,
	base string,
) (CheckView, []spider.Signal, error) {
	result, err := w.checker.Check(ctx, action.CheckInput{
		Capability: capability,
		All:        all,
		Base:       base,
	})

	return checkView(result), signalViews(result.Signals), err
}

func checkView(result action.CheckResult) CheckView {
	view := CheckView{
		Base:         result.Base,
		Scope:        result.Scope,
		Capabilities: nonNil(result.Capabilities),
		Promises:     promiseResults(result.Promises),
		Invocations:  make([]InvocationView, 0, len(result.Invocations)),
		Manual:       nonNil(result.Manual),
		Failed:       promiseResults(result.Failed),
	}

	for _, invocation := range result.Invocations {
		view.Invocations = append(view.Invocations, InvocationView(invocation))
	}

	return view
}

func promiseResults(reports []action.PromiseReport) []PromiseResultView {
	views := make([]PromiseResultView, 0, len(reports))

	for _, report := range reports {
		views = append(views, PromiseResultView{
			ID:         report.ID,
			Capability: report.Capability,
			Status:     report.Status,
			Reason:     report.Reason,
			State:      report.State,
			Drifted:    report.Drifted,
		})
	}

	return views
}
