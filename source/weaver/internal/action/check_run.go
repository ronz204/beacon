package action

import (
	"context"
	"errors"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
)

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
