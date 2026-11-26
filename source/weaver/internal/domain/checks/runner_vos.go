package checks

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/source/kernel"
)

var runnerNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

var paramNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)

var placeholderPattern = regexp.MustCompile(`\{([^{}]*)\}`)

type RunnerName struct {
	value string
}

func ParseRunnerName(raw string) (RunnerName, error) {
	value := strings.TrimSpace(raw)

	if !runnerNamePattern.MatchString(value) {
		return RunnerName{}, invalidRunnerValue(RunnerNameInvalid, raw)
	}

	return RunnerName{value: value}, nil
}

func (n RunnerName) String() string {
	return n.value
}

type PromiseRef struct {
	value kernel.RecordID
}

func ParsePromiseRef(raw string) (PromiseRef, error) {
	id, err := kernel.ParseRootRecordID(raw)
	if err != nil {
		return PromiseRef{}, invalidRunnerValue(RunnerPromiseBad, raw)
	}

	return PromiseRef{value: id}, nil
}

func (r PromiseRef) String() string {
	return r.value.String()
}

type RunnerSpecInput struct {
	Name string
	Run  []string
}

type RunnerSpec struct {
	name RunnerName
	run  []string
}

func ParseRunnerSpec(input RunnerSpecInput) (RunnerSpec, error) {
	name, err := ParseRunnerName(input.Name)
	if err != nil {
		return RunnerSpec{}, err
	}

	if len(input.Run) == 0 {
		return RunnerSpec{}, invalidRunnerValue(RunnerRunInvalid, "")
	}

	for _, item := range input.Run {
		if strings.TrimSpace(item) == "" {
			return RunnerSpec{}, invalidRunnerValue(RunnerRunInvalid, item)
		}
	}

	if placeholderPattern.MatchString(input.Run[0]) {
		return RunnerSpec{}, invalidRunnerValue(RunnerRunInvalid, input.Run[0])
	}

	return RunnerSpec{name: name, run: slices.Clone(input.Run)}, nil
}

func (s RunnerSpec) Name() RunnerName {
	return s.name
}

func (s RunnerSpec) Run() []string {
	return slices.Clone(s.run)
}

type Invocation struct {
	runner RunnerName
	argv   []string
}

func (i Invocation) Runner() RunnerName {
	return i.runner
}

func (i Invocation) Argv() []string {
	return slices.Clone(i.argv)
}

type PromiseCheckInput struct {
	Promise string
	Runner  string
	Params  map[string]string
}

type PromiseCheck struct {
	promise PromiseRef
	runner  RunnerName
	params  map[string]string
}

func ParsePromiseCheck(input PromiseCheckInput) (PromiseCheck, error) {
	promise, err := ParsePromiseRef(input.Promise)
	if err != nil {
		return PromiseCheck{}, err
	}

	runner, err := ParseRunnerName(input.Runner)
	if err != nil {
		return PromiseCheck{}, err
	}

	for key, value := range input.Params {
		if !paramNamePattern.MatchString(key) {
			return PromiseCheck{}, invalidRunnerValue(ParamNameInvalid, key)
		}
		if !isPlainParam(value) {
			return PromiseCheck{}, invalidRunnerValue(ParamInvalid, value)
		}
	}

	return PromiseCheck{
		promise: promise,
		runner:  runner,
		params:  maps.Clone(input.Params),
	}, nil
}

func (c PromiseCheck) Promise() PromiseRef {
	return c.promise
}

func isPlainParam(value string) bool {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" || strings.HasPrefix(trimmed, "-") {
		return false
	}

	return strings.IndexFunc(value, unicode.IsControl) < 0
}

type InvocationGroup struct {
	runner   RunnerName
	params   map[string]string
	promises []PromiseRef
}

func (g InvocationGroup) Runner() RunnerName {
	return g.runner
}

func (g InvocationGroup) Params() map[string]string {
	return maps.Clone(g.params)
}

func (g InvocationGroup) Promises() []PromiseRef {
	return slices.Clone(g.promises)
}
