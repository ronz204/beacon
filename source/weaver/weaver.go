package weaver

import (
	"context"
	"io"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
	"github.com/ronz204/beacon/source/weaver/internal/driver/branch"
	"github.com/ronz204/beacon/source/weaver/internal/driver/facade"
	"github.com/ronz204/beacon/source/weaver/internal/driver/identity"
	"github.com/ronz204/beacon/source/weaver/internal/driver/runner"
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
	Guidance  []string     `json:"guidance,omitempty"`
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

type PromiseResultView struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
	State      string `json:"state"`
	Drifted    bool   `json:"drifted"`
}

type InvocationView struct {
	Runner   string   `json:"runner"`
	Argv     []string `json:"argv"`
	Promises []string `json:"promises"`
	Status   string   `json:"status"`
	Reason   string   `json:"reason,omitempty"`
	Output   string   `json:"output,omitempty"`
}

type CheckView struct {
	Base         string              `json:"base,omitempty"`
	Scope        string              `json:"scope"`
	Capabilities []string            `json:"capabilities"`
	Promises     []PromiseResultView `json:"promises"`
	Invocations  []InvocationView    `json:"invocations"`
	Manual       []string            `json:"manual"`
	Failed       []PromiseResultView `json:"failed"`
}

type EntryView struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	Target   string `json:"target"`
	ID       string `json:"id"`
	Level    int    `json:"level"`
	Reason   string `json:"reason,omitempty"`
	Before   string `json:"before,omitempty"`
	After    string `json:"after,omitempty"`
	NeedsYes bool   `json:"needsYes"`
	Approved bool   `json:"approved"`
}

type LevelView struct {
	Level   int         `json:"level"`
	Changes []EntryView `json:"changes"`
}

type DiffView struct {
	Revision string      `json:"revision"`
	Levels   []LevelView `json:"levels"`
	Pending  []EntryView `json:"pending"`
}

type Weaver struct {
	project spider.Project
	changes deltas.ChangeSource
	records action.Recorder
	closer  action.Closer
	claimer action.Claimer
	checker action.Checker
	differ  action.Differ
}

func New(project spider.Project) Weaver {
	today, _ := kernel.CalendarDateOf(time.Now())
	records := facade.NewRecords(project)
	who := identity.Git{}

	differ := action.Differ{
		Changes: facade.NewChanges(project),
		Base:    branch.Git{},
	}

	return Weaver{
		project: project,
		changes: facade.NewChanges(project),
		records: action.Recorder{
			Store:    records,
			Identity: who,
			Today:    today,
			Pending:  differ,
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
		differ: differ,
		checker: action.Checker{
			Source:   facade.NewChecks(project),
			Base:     branch.Git{},
			Executor: runner.New(nil),
		},
	}
}

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

type Summary struct {
	View           spider.BriefView
	Findings       []spider.Finding
	PendingChecked bool
}

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
		views = append(views, PromiseResultView(report))
	}

	return views
}

func signalViews(signals []action.Signal) []spider.Signal {
	views := make([]spider.Signal, 0, len(signals))

	for _, signal := range signals {
		views = append(views, spider.Signal(signal))
	}

	return views
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}

	return values
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
