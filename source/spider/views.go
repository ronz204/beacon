package spider

import (
	"encoding/json"
	"fmt"

	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

type Record struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"record"`
}

type StoredRecord struct {
	Record
	File     string `json:"file"`
	Archived bool   `json:"archived"`
}

type Signal struct {
	Kind     string `json:"kind"`
	Record   string `json:"record"`
	Related  string `json:"related,omitempty"`
	Relation string `json:"relation,omitempty"`
}

type Summary struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type BriefLine struct {
	Section string `json:"section"`
	Text    string `json:"text"`
}

type BriefView struct {
	Lines   []BriefLine `json:"lines"`
	Omitted int         `json:"omitted"`
}

type PromiseView struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Text       string `json:"text"`
	Status     string `json:"status"`
	State      string `json:"state"`
	NotRunHere bool   `json:"notRunHere"`
	Drifted    bool   `json:"drifted"`
}

type CapabilityView struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Promises []PromiseView `json:"promises"`
}

type ContextView struct {
	Goals        []Summary        `json:"goals"`
	Capabilities []CapabilityView `json:"capabilities"`
	Decisions    []Summary        `json:"decisions"`
	Guides       []Summary        `json:"guides"`
	Questions    []Summary        `json:"questions"`
	Signals      []Signal         `json:"signals"`
}

type HookView struct {
	Lines   []string `json:"lines"`
	Omitted int      `json:"omitted"`
}

type ReadyView struct {
	Works []Summary `json:"works"`
	Tasks []Summary `json:"tasks"`
}

type AnalysisView struct {
	Promises []PromiseView `json:"promises"`
	Signals  []Signal      `json:"signals"`
}

type Match struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Text string `json:"text"`
}

type FindView struct {
	Matches []Match `json:"matches"`
	Omitted int     `json:"omitted"`
}

type Closure struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Approved string `json:"approved,omitempty"`
}

type StateChange struct {
	Path     string    `json:"path"`
	Kind     string    `json:"kind"`
	Baseline bool      `json:"baseline"`
	Runners  []string  `json:"runners,omitempty"`
	Closed   []Closure `json:"closed,omitempty"`
	Before   *Record   `json:"before"`
	After    *Record   `json:"after"`
}

func summaryOf(id, text fmt.Stringer) Summary {
	return Summary{ID: id.String(), Text: firstLine(text.String())}
}

func signalViews(signals []thread.Signal) []Signal {
	views := make([]Signal, 0, len(signals))

	for _, signal := range signals {
		views = append(views, Signal{
			Kind:     string(signal.Kind),
			Record:   signal.Subject.String(),
			Related:  signal.Related.String(),
			Relation: string(signal.Relation),
		})
	}

	return views
}

func briefView(result action.BriefResult) BriefView {
	lines := make([]BriefLine, 0, len(result.Lines))

	for _, line := range result.Lines {
		lines = append(lines, BriefLine{
			Section: string(line.Section),
			Text:    line.Text,
		})
	}

	return BriefView{Lines: lines, Omitted: result.Omitted}
}

func promiseView(
	capability ledger.Capability,
	promise ledger.Promise,
	verification ledger.Verification,
) PromiseView {
	return PromiseView{
		ID:         promise.ID().String(),
		Capability: capability.ID().String(),
		Text:       promise.Text().String(),
		Status:     string(promise.Status()),
		State:      string(verification.Status()),
		NotRunHere: verification.NotRunHere(),
		Drifted:    verification.Drifted(),
	}
}

func contextView(result action.ContextResult) ContextView {
	view := ContextView{
		Goals:        []Summary{},
		Capabilities: []CapabilityView{},
		Decisions:    []Summary{},
		Guides:       []Summary{},
		Questions:    []Summary{},
		Signals:      signalViews(result.Signals),
	}

	for _, goal := range result.Goals {
		view.Goals = append(view.Goals, summaryOf(goal.ID(), goal.Statement()))
	}
	for _, entry := range result.Capabilities {
		view.Capabilities = append(view.Capabilities, capabilityView(entry))
	}
	for _, decision := range result.Decisions {
		view.Decisions = append(
			view.Decisions,
			summaryOf(decision.ID(), decision.What()),
		)
	}
	for _, guide := range result.Guides {
		view.Guides = append(view.Guides, summaryOf(guide.ID(), guide.Title()))
	}
	for _, question := range result.Questions {
		view.Questions = append(
			view.Questions,
			summaryOf(question.ID(), question.Text()),
		)
	}

	return view
}

func capabilityView(entry action.CapabilityContext) CapabilityView {
	view := CapabilityView{
		ID:       entry.Capability.ID().String(),
		Name:     entry.Capability.Name().String(),
		Promises: make([]PromiseView, 0, len(entry.Promises)),
	}

	for _, promise := range entry.Promises {
		view.Promises = append(view.Promises, promiseView(
			entry.Capability,
			promise.Promise,
			promise.Verification,
		))
	}

	return view
}

func readyView(result action.ReadyResult) ReadyView {
	view := ReadyView{
		Works: make([]Summary, 0, len(result.Works)),
		Tasks: make([]Summary, 0, len(result.Tasks)),
	}

	for _, work := range result.Works {
		view.Works = append(view.Works, summaryOf(work.ID(), work.Request()))
	}
	for _, task := range result.Tasks {
		view.Tasks = append(view.Tasks, taskSummary(task))
	}

	return view
}

func taskSummary(task agenda.Task) Summary {
	return summaryOf(task.ID(), task.Title())
}

func analysisView(analysis action.Analysis) AnalysisView {
	view := AnalysisView{
		Promises: []PromiseView{},
		Signals:  signalViews(analysis.Signals),
	}

	for _, capability := range analysis.Snapshot.Capabilities {
		for _, promise := range capability.Promises() {
			verification := analysis.Verifications[promise.ID()]
			view.Promises = append(
				view.Promises,
				promiseView(capability, promise, verification),
			)
		}
	}

	return view
}

func findView(result action.FindResult) FindView {
	view := FindView{
		Matches: make([]Match, 0, len(result.Matches)),
		Omitted: result.Omitted,
	}

	for _, match := range result.Matches {
		view.Matches = append(view.Matches, Match{
			Type: string(match.Kind),
			ID:   match.ID,
			Text: match.Text,
		})
	}

	return view
}

type RunnerConfig struct {
	Name     string   `json:"name"`
	Run      []string `json:"run"`
	Baseline []string `json:"baseline"`
}
