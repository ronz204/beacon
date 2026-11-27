package action

import (
	"context"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
)

const keptOutputLines = 40

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
