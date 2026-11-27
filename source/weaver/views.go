package weaver

import (
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
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

type ManualView struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Text       string `json:"text"`
	Manual     string `json:"manual"`
	State      string `json:"state"`
	Drifted    bool   `json:"drifted"`
}

type AttestView struct {
	Pending  []ManualView `json:"pending"`
	Attested []string     `json:"attested"`
	Commit   string       `json:"commit,omitempty"`
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

type RiskView struct {
	Level   string   `json:"level"`
	Reasons []string `json:"reasons"`
}

type DiffView struct {
	Revision string      `json:"revision"`
	Risk     RiskView    `json:"risk"`
	Levels   []LevelView `json:"levels"`
	Pending  []EntryView `json:"pending"`
}

type ReviewPromiseView struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Text       string `json:"text"`
	State      string `json:"state"`
	Drifted    bool   `json:"drifted"`
}

type ReviewView struct {
	Promises []ReviewPromiseView `json:"promises"`
	Changes  []EntryView         `json:"changes"`
	Rules    []string            `json:"rules"`
}

type VerifyView struct {
	Revision string      `json:"revision"`
	Risk     RiskView    `json:"risk"`
	Check    CheckView   `json:"check"`
	Review   *ReviewView `json:"review,omitempty"`
}

type Summary struct {
	View           spider.BriefView
	Findings       []spider.Finding
	PendingChecked bool
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
