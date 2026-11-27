package action

import (
	"context"
)

type RunnerDef struct {
	Name string
	Run  []string
}

type PromiseDef struct {
	ID     string
	Status string
	Runner string
	Params map[string]string
	Manual bool
}

type TargetDef struct {
	ID       string
	Promises []PromiseDef
}

type Verdict struct {
	Promise string
	Outcome string
	Reason  string
}

type Signal struct {
	Kind     string
	Record   string
	Related  string
	Relation string
}

type PromiseState struct {
	ID      string
	Text    string
	State   string
	Drifted bool
}

type Analysis struct {
	Promises []PromiseState
	Signals  []Signal
}

type CheckSource interface {
	Root() string
	Runners(ctx context.Context) ([]RunnerDef, error)
	RunnersAt(ctx context.Context, revision string) ([]RunnerDef, error)
	Targets(ctx context.Context) ([]TargetDef, error)
	Touched(ctx context.Context, revision string) ([]string, error)
	Record(ctx context.Context, verdicts []Verdict) error
	Analyze(ctx context.Context) (Analysis, error)
}

type BaseResolver interface {
	Base(ctx context.Context, forced string) (string, error)
}
