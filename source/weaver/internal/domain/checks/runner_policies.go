package checks

import (
	"maps"
	"slices"
	"strings"
)

func BuildInvocation(
	spec RunnerSpec,
	params map[string]string,
) (Invocation, error) {
	used := map[string]bool{}
	argv := make([]string, 0, len(spec.run))

	for _, item := range spec.run {
		built, err := substitute(spec, item, params, used)
		if err != nil {
			return Invocation{}, err
		}

		argv = append(argv, built)
	}

	for _, name := range slices.Sorted(maps.Keys(params)) {
		if !used[name] {
			return Invocation{}, runnerErrors.Fail(ParamUnused, map[string]any{
				"runner": spec.name.String(),
				"param":  name,
			})
		}
	}

	return Invocation{runner: spec.name, argv: argv}, nil
}

func substitute(
	spec RunnerSpec,
	item string,
	params map[string]string,
	used map[string]bool,
) (string, error) {
	var failure error

	fill := func(raw string) string {
		name := strings.Trim(raw, "{}")
		value, found := params[name]

		switch {
		case failure != nil:
			return raw
		case !found:
			failure = missingParam(spec, name)
		case !isPlainParam(value):
			failure = invalidRunnerValue(ParamInvalid, value)
		}
		used[name] = true

		return value
	}

	built := placeholderPattern.ReplaceAllStringFunc(item, fill)

	return built, failure
}

func missingParam(spec RunnerSpec, name string) error {
	return runnerErrors.Fail(ParamMissing, map[string]any{
		"runner": spec.name.String(),
		"param":  name,
	})
}

func GroupInvocations(checks []PromiseCheck) []InvocationGroup {
	ordered := slices.Clone(checks)
	slices.SortStableFunc(ordered, func(a, b PromiseCheck) int {
		return strings.Compare(a.promise.String(), b.promise.String())
	})

	var groups []InvocationGroup
	position := map[string]int{}

	for _, check := range ordered {
		key := invocationKey(check)
		index, seen := position[key]

		if !seen {
			position[key] = len(groups)
			groups = append(groups, InvocationGroup{
				runner: check.runner,
				params: maps.Clone(check.params),
			})
			index = len(groups) - 1
		}

		groups[index].promises = append(groups[index].promises, check.promise)
	}

	return groups
}

func invocationKey(check PromiseCheck) string {
	var key strings.Builder
	key.WriteString(check.runner.String())

	for _, name := range slices.Sorted(maps.Keys(check.params)) {
		key.WriteByte(0)
		key.WriteString(name)
		key.WriteByte('=')
		key.WriteString(check.params[name])
	}

	return key.String()
}
