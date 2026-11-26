package checks_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
)

func sampleSpec(t *testing.T) checks.RunnerSpec {
	t.Helper()

	spec, err := checks.ParseRunnerSpec(checks.RunnerSpecInput{
		Name: "test-node",
		Run:  []string{"pnpm", "--filter", "{pkg}", "test", "{target}"},
	})
	if err != nil {
		t.Fatal(err)
	}

	return spec
}

func codeOf(t *testing.T, err error) checks.RunnerCode {
	t.Helper()

	var failure checks.RunnerError
	if !errors.As(err, &failure) {
		t.Fatalf("err = %v", err)
	}

	return failure.Code
}

func TestBuildInvocationSubstitutesEachItemWithoutAShell(t *testing.T) {
	params := map[string]string{"pkg": "api", "target": "auth/lockout"}

	invocation, err := checks.BuildInvocation(sampleSpec(t), params)

	if err != nil {
		t.Fatal(err)
	}

	want := []string{"pnpm", "--filter", "api", "test", "auth/lockout"}
	if !slices.Equal(invocation.Argv(), want) {
		t.Fatalf("argv = %q", invocation.Argv())
	}
}

func TestBuildInvocationKeepsASpacedValueAsOneItem(t *testing.T) {
	params := map[string]string{"pkg": "a b; rm -rf", "target": "x"}

	invocation, err := checks.BuildInvocation(sampleSpec(t), params)

	if err != nil {
		t.Fatal(err)
	}
	if invocation.Argv()[2] != "a b; rm -rf" {
		t.Fatalf("argv = %q", invocation.Argv())
	}
}

func TestBuildInvocationFailsWhenAParamIsMissing(t *testing.T) {
	_, err := checks.BuildInvocation(
		sampleSpec(t),
		map[string]string{"pkg": "api"},
	)

	if codeOf(t, err) != checks.ParamMissing {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildInvocationFailsWhenAParamIsUnused(t *testing.T) {
	params := map[string]string{"pkg": "a", "target": "b", "targt": "c"}

	_, err := checks.BuildInvocation(sampleSpec(t), params)

	if codeOf(t, err) != checks.ParamUnused {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildInvocationRejectsAParamThatLooksLikeAnOption(t *testing.T) {
	for _, value := range []string{"--config=evil", "-x", " -y", ""} {
		params := map[string]string{"pkg": value, "target": "b"}

		_, err := checks.BuildInvocation(sampleSpec(t), params)

		if codeOf(t, err) != checks.ParamInvalid {
			t.Fatalf("value %q: err = %v", value, err)
		}
	}
}

func TestParseRunnerSpecRejectsAParamInTheProgram(t *testing.T) {
	_, err := checks.ParseRunnerSpec(checks.RunnerSpecInput{
		Name: "evil",
		Run:  []string{"{pkg}", "test"},
	})

	if codeOf(t, err) != checks.RunnerRunInvalid {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRunnerSpecRejectsEmptyRun(t *testing.T) {
	_, err := checks.ParseRunnerSpec(checks.RunnerSpecInput{Name: "empty"})

	if codeOf(t, err) != checks.RunnerRunInvalid {
		t.Fatalf("err = %v", err)
	}
}

func checkOf(
	t *testing.T,
	promise string,
	params map[string]string,
) checks.PromiseCheck {
	t.Helper()

	parsed, err := checks.ParsePromiseCheck(checks.PromiseCheckInput{
		Promise: promise,
		Runner:  "test-node",
		Params:  params,
	})
	if err != nil {
		t.Fatal(err)
	}

	return parsed
}

func TestGroupInvocationsSharesTheSameCheck(t *testing.T) {
	shared := map[string]string{"pkg": "api", "target": "auth"}
	other := map[string]string{"pkg": "web", "target": "auth"}

	groups := checks.GroupInvocations([]checks.PromiseCheck{
		checkOf(t, "bc-d5e6", shared),
		checkOf(t, "bc-a1b2", other),
		checkOf(t, "bc-c3d4", shared),
	})

	if len(groups) != 2 {
		t.Fatalf("groups = %d", len(groups))
	}

	first := groups[0].Promises()
	if len(first) != 1 || first[0].String() != "bc-a1b2" {
		t.Fatalf("first = %v", first)
	}

	second := groups[1].Promises()
	if len(second) != 2 || second[0].String() != "bc-c3d4" {
		t.Fatalf("second = %v", second)
	}
}

func TestParsePromiseCheckRejectsAnOptionLikeParam(t *testing.T) {
	_, err := checks.ParsePromiseCheck(checks.PromiseCheckInput{
		Promise: "bc-c3d4",
		Runner:  "test-node",
		Params:  map[string]string{"pkg": "--evil"},
	})

	if codeOf(t, err) != checks.ParamInvalid {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildInvocationLeavesLiteralBracesAlone(t *testing.T) {
	spec, err := checks.ParseRunnerSpec(checks.RunnerSpecInput{
		Name: "emit",
		Run:  []string{"emit", `{"mode":"fast"}`, "{target}"},
	})
	if err != nil {
		t.Fatal(err)
	}

	invocation, err := checks.BuildInvocation(
		spec,
		map[string]string{"target": "auth"},
	)

	if err != nil {
		t.Fatal(err)
	}
	if invocation.Argv()[1] != `{"mode":"fast"}` {
		t.Fatalf("argv = %q", invocation.Argv())
	}
}
