package action

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
)

const keptOutputLines = 40

type CheckCode string

const (
	BaseMissing       CheckCode = "base_missing"
	RunnerMissing     CheckCode = "runner_missing"
	CapabilityMissing CheckCode = "capability_missing"
)

type CheckError = kernel.Error[CheckCode]

var checkErrors = kernel.ErrorCatalog[CheckCode]{
	BaseMissing:       "no base branch was found to compare against",
	RunnerMissing:     "a check names a runner that config.yaml does not have",
	CapabilityMissing: "the id is not a capability",
}

func FailBaseMissing(base string) CheckError {
	return checkErrors.Fail(BaseMissing, map[string]any{"base": base})
}

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
	Targets(ctx context.Context) ([]TargetDef, error)
	Touched(ctx context.Context, revision string) ([]string, error)
	Record(ctx context.Context, verdicts []Verdict) error
	Analyze(ctx context.Context) (Analysis, error)
}

type BaseResolver interface {
	Base(ctx context.Context, forced string) (string, error)
}

type CheckInput struct {
	Capability string
	All        bool
	Base       string
}

type PromiseReport struct {
	ID         string
	Capability string
	Status     string
	Reason     string
	State      string
	Drifted    bool
}

type InvocationReport struct {
	Runner   string
	Argv     []string
	Promises []string
	Status   string
	Reason   string
	Output   string
}

type CheckResult struct {
	Base         string
	Scope        string
	Capabilities []string
	Promises     []PromiseReport
	Invocations  []InvocationReport
	Manual       []string
	Failed       []PromiseReport
	Signals      []Signal
}

type Checker struct {
	Source   CheckSource
	Base     BaseResolver
	Executor checks.Executor
}

type plannedGroup struct {
	invocation checks.Invocation
	group      checks.InvocationGroup
}

func (c Checker) Check(
	ctx context.Context,
	input CheckInput,
) (CheckResult, error) {
	targets, err := c.Source.Targets(ctx)
	if err != nil {
		return CheckResult{}, err
	}

	result := CheckResult{Scope: scopeOf(input)}
	selected, err := c.choose(ctx, input, targets, &result)
	if err != nil {
		return result, err
	}

	result.Capabilities = idsOf(selected)

	owners := map[string]string{}
	var promises []PromiseDef
	for _, target := range selected {
		for _, promise := range target.Promises {
			owners[promise.ID] = target.ID
			if promise.Manual {
				result.Manual = append(result.Manual, promise.ID)
				continue
			}

			promises = append(promises, promise)
		}
	}

	planned, err := c.plan(ctx, promises)
	if err != nil {
		return result, err
	}

	verdicts, err := c.run(ctx, planned, &result)
	if err != nil {
		return result, err
	}

	return c.finish(ctx, result, verdicts, owners)
}

func scopeOf(input CheckInput) string {
	switch {
	case input.Capability != "":
		return "capability"
	case input.All:
		return "all"
	}

	return "changed"
}

func (c Checker) choose(
	ctx context.Context,
	input CheckInput,
	targets []TargetDef,
	result *CheckResult,
) ([]TargetDef, error) {
	switch {
	case input.Capability != "":
		return oneTarget(targets, input.Capability)
	case input.All:
		return targets, nil
	}

	base, err := c.Base.Base(ctx, input.Base)
	if err != nil {
		return nil, err
	}

	result.Base = base
	touched, err := c.Source.Touched(ctx, base)
	if err != nil {
		return nil, err
	}

	return filterTargets(targets, touched), nil
}

func oneTarget(targets []TargetDef, id string) ([]TargetDef, error) {
	for _, target := range targets {
		if target.ID == id {
			return []TargetDef{target}, nil
		}
	}

	return nil, checkErrors.Fail(CapabilityMissing, map[string]any{"id": id})
}

func filterTargets(targets []TargetDef, ids []string) []TargetDef {
	var selected []TargetDef

	for _, target := range targets {
		if slices.Contains(ids, target.ID) {
			selected = append(selected, target)
		}
	}

	return selected
}

func idsOf(targets []TargetDef) []string {
	ids := make([]string, 0, len(targets))

	for _, target := range targets {
		ids = append(ids, target.ID)
	}

	return ids
}

func (c Checker) plan(
	ctx context.Context,
	promises []PromiseDef,
) ([]plannedGroup, error) {
	defs, err := c.Source.Runners(ctx)
	if err != nil {
		return nil, err
	}

	specs, err := specsOf(defs)
	if err != nil {
		return nil, err
	}

	parsed := make([]checks.PromiseCheck, 0, len(promises))
	for _, promise := range promises {
		if _, found := specs[promise.Runner]; !found {
			return nil, checkErrors.Fail(RunnerMissing, map[string]any{
				"runner":  promise.Runner,
				"promise": promise.ID,
			})
		}

		check, err := checks.ParsePromiseCheck(checks.PromiseCheckInput{
			Promise: promise.ID,
			Runner:  promise.Runner,
			Params:  promise.Params,
		})
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, check)
	}

	return buildGroups(specs, checks.GroupInvocations(parsed))
}

func specsOf(defs []RunnerDef) (map[string]checks.RunnerSpec, error) {
	specs := make(map[string]checks.RunnerSpec, len(defs))

	for _, def := range defs {
		spec, err := checks.ParseRunnerSpec(checks.RunnerSpecInput{
			Name: def.Name,
			Run:  def.Run,
		})
		if err != nil {
			return nil, err
		}

		specs[def.Name] = spec
	}

	return specs, nil
}

func buildGroups(
	specs map[string]checks.RunnerSpec,
	groups []checks.InvocationGroup,
) ([]plannedGroup, error) {
	planned := make([]plannedGroup, 0, len(groups))

	for _, group := range groups {
		spec := specs[group.Runner().String()]

		invocation, err := checks.BuildInvocation(spec, group.Params())
		if err != nil {
			return nil, err
		}

		planned = append(planned, plannedGroup{invocation, group})
	}

	return planned, nil
}

func (c Checker) run(
	ctx context.Context,
	planned []plannedGroup,
	result *CheckResult,
) ([]Verdict, error) {
	var verdicts []Verdict

	for _, step := range planned {
		report := InvocationReport{
			Runner:   step.invocation.Runner().String(),
			Argv:     step.invocation.Argv(),
			Promises: promiseIDs(step.group),
		}

		execution, err := c.Executor.Run(
			ctx,
			step.invocation.Argv(),
			c.Source.Root(),
		)
		if unavailable(err) {
			report.Status = "unavailable"
			report.Reason = string(checks.RunnerUnavailable)
			result.Invocations = append(result.Invocations, report)
			continue
		}
		if err != nil {
			return nil, err
		}

		outcome := checks.OutcomeOf(execution)
		report.Status = string(outcome.Status())
		report.Reason = string(outcome.Reason())
		if !outcome.Passed() {
			report.Output = lastLines(execution.Output, keptOutputLines)
		}

		result.Invocations = append(result.Invocations, report)
		verdicts = append(verdicts, verdictsOf(report)...)
	}

	return verdicts, nil
}

func unavailable(err error) bool {
	var failure checks.RunnerError

	return errors.As(err, &failure) && failure.Code == checks.RunnerUnavailable
}

func promiseIDs(group checks.InvocationGroup) []string {
	refs := group.Promises()
	ids := make([]string, 0, len(refs))

	for _, ref := range refs {
		ids = append(ids, ref.String())
	}

	return ids
}

func verdictsOf(report InvocationReport) []Verdict {
	verdicts := make([]Verdict, 0, len(report.Promises))

	for _, promise := range report.Promises {
		verdicts = append(verdicts, Verdict{
			Promise: promise,
			Outcome: report.Status,
			Reason:  report.Reason,
		})
	}

	return verdicts
}

func (c Checker) finish(
	ctx context.Context,
	result CheckResult,
	verdicts []Verdict,
	owners map[string]string,
) (CheckResult, error) {
	if len(verdicts) > 0 {
		if err := c.Source.Record(ctx, verdicts); err != nil {
			return result, err
		}
	}

	analysis, err := c.Source.Analyze(ctx)
	if err != nil {
		return result, err
	}

	result.Signals = analysis.Signals
	result.Promises = promiseReports(result, analysis, owners)

	for _, report := range result.Promises {
		if report.Status == string(checks.OutcomeFailed) {
			result.Failed = append(result.Failed, report)
		}
	}

	return result, nil
}

func promiseReports(
	result CheckResult,
	analysis Analysis,
	owners map[string]string,
) []PromiseReport {
	states := map[string]PromiseState{}
	for _, state := range analysis.Promises {
		states[state.ID] = state
	}

	var reports []PromiseReport
	for _, invocation := range result.Invocations {
		for _, id := range invocation.Promises {
			reports = append(reports, PromiseReport{
				ID:         id,
				Capability: owners[id],
				Status:     invocation.Status,
				Reason:     invocation.Reason,
				State:      states[id].State,
				Drifted:    states[id].Drifted,
			})
		}
	}

	slices.SortFunc(reports, func(a, b PromiseReport) int {
		return strings.Compare(a.ID, b.ID)
	})

	return reports
}

func lastLines(text string, count int) string {
	lines := strings.Split(strings.TrimRight(text, "\r\n"), "\n")
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}

	return strings.Join(lines, "\n")
}
