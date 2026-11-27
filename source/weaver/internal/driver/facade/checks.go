package facade

import (
	"context"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

type Checks struct {
	project spider.Project
}

func NewChecks(project spider.Project) Checks {
	return Checks{project: project}
}

func (c Checks) Root() string {
	return c.project.Worktree()
}

func (c Checks) Runners(ctx context.Context) ([]action.RunnerDef, error) {
	runners, findings := c.project.Runners()
	if hasErrors(findings) {
		return nil, action.FailStateInvalid(findings)
	}

	defs := make([]action.RunnerDef, 0, len(runners))
	for _, runner := range runners {
		defs = append(defs, action.RunnerDef{
			Name: runner.Name,
			Run:  runner.Run,
		})
	}

	return defs, nil
}

func (c Checks) RunnersAt(
	ctx context.Context,
	revision string,
) ([]action.RunnerDef, error) {
	runners, err := c.project.RunnersAt(ctx, revision)
	if err != nil {
		return nil, err
	}

	defs := make([]action.RunnerDef, 0, len(runners))
	for _, runner := range runners {
		defs = append(defs, action.RunnerDef{
			Name: runner.Name,
			Run:  runner.Run,
		})
	}

	return defs, nil
}

func (c Checks) Targets(ctx context.Context) ([]action.TargetDef, error) {
	targets, findings := c.project.CheckTargets(ctx)
	if hasErrors(findings) {
		return nil, action.FailStateInvalid(findings)
	}

	defs := make([]action.TargetDef, 0, len(targets))
	for _, target := range targets {
		def := action.TargetDef{ID: target.ID}
		for _, promise := range target.Promises {
			def.Promises = append(def.Promises, action.PromiseDef{
				ID:     promise.ID,
				Status: promise.Status,
				Runner: promise.Runner,
				Params: promise.Params,
				Manual: promise.Manual != "",
			})
		}

		defs = append(defs, def)
	}

	return defs, nil
}

func (c Checks) Touched(
	ctx context.Context,
	revision string,
) ([]string, error) {
	return c.project.TouchedCapabilities(ctx, revision)
}

func (c Checks) Record(
	ctx context.Context,
	verdicts []action.Verdict,
) error {
	converted := make([]spider.Verdict, 0, len(verdicts))
	for _, verdict := range verdicts {
		converted = append(converted, spider.Verdict(verdict))
	}

	return c.project.RecordResults(ctx, converted)
}

func (c Checks) Analyze(ctx context.Context) (action.Analysis, error) {
	view, findings := c.project.Analyze(ctx)
	if hasErrors(findings) {
		return action.Analysis{}, action.FailStateInvalid(findings)
	}

	analysis := action.Analysis{Signals: signalsOf(view.Signals)}
	for _, promise := range view.Promises {
		analysis.Promises = append(analysis.Promises, action.PromiseState{
			ID:      promise.ID,
			Text:    promise.Text,
			State:   promise.State,
			Drifted: promise.Drifted,
		})
	}

	return analysis, nil
}

func signalsOf(signals []spider.Signal) []action.Signal {
	converted := make([]action.Signal, 0, len(signals))

	for _, signal := range signals {
		converted = append(converted, action.Signal(signal))
	}

	return converted
}

func hasErrors(findings []spider.Finding) bool {
	for _, finding := range findings {
		if finding.Severity == spider.FindingError {
			return true
		}
	}

	return false
}
