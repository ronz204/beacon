package action

import (
	"context"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
)

type plannedGroup struct {
	invocation checks.Invocation
	group      checks.InvocationGroup
}

func (c Checker) runners(
	ctx context.Context,
	input CheckInput,
) ([]RunnerDef, error) {
	if input.RunnersAt != "" {
		return c.Source.RunnersAt(ctx, input.RunnersAt)
	}

	return c.Source.Runners(ctx)
}

func (c Checker) plan(
	ctx context.Context,
	input CheckInput,
	promises []PromiseDef,
	result *CheckResult,
) ([]plannedGroup, error) {
	defs, err := c.runners(ctx, input)
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
			if input.RunnersAt != "" {
				result.Skipped = append(result.Skipped, promise.ID)
				continue
			}

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
