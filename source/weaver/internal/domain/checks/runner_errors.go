package checks

import "github.com/ronz204/beacon/source/kernel"

type RunnerCode string

const (
	RunnerNameInvalid RunnerCode = "runner_name_invalid"
	RunnerRunInvalid  RunnerCode = "runner_run_invalid"
	RunnerUnavailable RunnerCode = "runner_unavailable"
	RunnerPromiseBad  RunnerCode = "runner_promise_invalid"
	ParamMissing      RunnerCode = "param_missing"
	ParamUnused       RunnerCode = "param_unused"
	ParamInvalid      RunnerCode = "param_invalid"
	ParamNameInvalid  RunnerCode = "param_name_invalid"
)

type RunnerError = kernel.Error[RunnerCode]

var runnerErrors = kernel.ErrorCatalog[RunnerCode]{
	RunnerNameInvalid: "runner name must be letters, digits, . _ -",
	RunnerRunInvalid:  "runner run needs a fixed program and no blank items",
	RunnerUnavailable: "the runner program could not be started",
	RunnerPromiseBad:  "promise id must be a root record id",
	ParamMissing:      "the runner uses a parameter the check does not set",
	ParamUnused:       "the check sets a parameter the runner does not use",
	ParamInvalid:      "a parameter must be a plain value, not an option",
	ParamNameInvalid:  "parameter name must be letters, digits, _",
}

func invalidRunnerValue(code RunnerCode, raw string) RunnerError {
	details := map[string]any{"value": raw}
	return runnerErrors.Fail(code, details)
}

func FailRunnerUnavailable(cause error) RunnerError {
	return runnerErrors.Fail(RunnerUnavailable, nil).WithCause(cause)
}
