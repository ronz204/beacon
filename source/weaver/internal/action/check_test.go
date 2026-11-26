package action_test

import (
	"context"
	"testing"

	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/checks"
	"github.com/ronz204/beacon/source/weaver/internal/driver/double"
)

type fakeSource struct {
	targets  []action.TargetDef
	recorded []action.Verdict
}

func (f *fakeSource) Root() string {
	return "/work/one"
}

func (f *fakeSource) Runners(context.Context) ([]action.RunnerDef, error) {
	return []action.RunnerDef{{
		Name: "test-go",
		Run:  []string{"go-test", "{target}"},
	}}, nil
}

func (f *fakeSource) Targets(context.Context) ([]action.TargetDef, error) {
	return f.targets, nil
}

func (f *fakeSource) Touched(context.Context, string) ([]string, error) {
	return nil, nil
}

func (f *fakeSource) Record(_ context.Context, v []action.Verdict) error {
	f.recorded = append(f.recorded, v...)
	return nil
}

func (f *fakeSource) Analyze(context.Context) (action.Analysis, error) {
	return action.Analysis{}, nil
}

type fakeBase struct{}

func (fakeBase) Base(context.Context, string) (string, error) {
	return "abc1234", nil
}

func promiseWith(id, target string) action.PromiseDef {
	return action.PromiseDef{
		ID:     id,
		Status: "active",
		Runner: "test-go",
		Params: map[string]string{"target": target},
	}
}

func sampleSource() *fakeSource {
	return &fakeSource{targets: []action.TargetDef{{
		ID: "bc-7f3a",
		Promises: []action.PromiseDef{
			promiseWith("bc-c3d4", "auth"),
			promiseWith("bc-c5b2", "auth"),
			promiseWith("bc-d6e7", "billing"),
			{ID: "bc-e8f9", Status: "active", Manual: true},
		},
	}}}
}

func TestCheckRunsOnceForPromisesSharingACheck(t *testing.T) {
	var runs []double.Run
	source := sampleSource()
	checker := action.Checker{
		Source: source,
		Base:   fakeBase{},
		Executor: double.Executor{
			Results: map[string]checks.ExecutionResult{
				"billing": {ExitCode: 1, Output: "billing broke"},
			},
			Runs: &runs,
		},
	}

	result, err := checker.Check(
		context.Background(),
		action.CheckInput{Capability: "bc-7f3a"},
	)

	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 2 {
		t.Fatalf("runs = %+v", runs)
	}
	if runs[0].Dir != "/work/one" {
		t.Fatalf("dir = %q", runs[0].Dir)
	}
	if len(source.recorded) != 3 {
		t.Fatalf("recorded = %+v", source.recorded)
	}
	if len(result.Manual) != 1 || len(result.Failed) != 1 {
		t.Fatalf("result = %+v", result)
	}
	if result.Failed[0].ID != "bc-d6e7" {
		t.Fatalf("failed = %+v", result.Failed)
	}
}

func TestCheckRecordsATimeoutAsAFailureWithItsReason(t *testing.T) {
	source := sampleSource()
	checker := action.Checker{
		Source: source,
		Base:   fakeBase{},
		Executor: double.Executor{Results: map[string]checks.ExecutionResult{
			"billing": {TimedOut: true},
		}},
	}

	_, err := checker.Check(
		context.Background(),
		action.CheckInput{Capability: "bc-7f3a"},
	)

	if err != nil {
		t.Fatal(err)
	}

	var timeout action.Verdict
	for _, verdict := range source.recorded {
		if verdict.Promise == "bc-d6e7" {
			timeout = verdict
		}
	}
	if timeout.Outcome != "failed" || timeout.Reason != "timeout" {
		t.Fatalf("verdict = %+v", timeout)
	}
}

func TestCheckRecordsNothingForAnUnavailableRunner(t *testing.T) {
	source := sampleSource()
	checker := action.Checker{
		Source: source,
		Base:   fakeBase{},
		Executor: double.Executor{
			Err: checks.FailRunnerUnavailable(context.Canceled),
		},
	}

	result, err := checker.Check(
		context.Background(),
		action.CheckInput{Capability: "bc-7f3a"},
	)

	if err != nil {
		t.Fatal(err)
	}
	if len(source.recorded) != 0 || len(result.Failed) != 0 {
		t.Fatalf("recorded = %+v, failed = %+v", source.recorded, result.Failed)
	}
	for _, invocation := range result.Invocations {
		if invocation.Status != "unavailable" {
			t.Fatalf("invocation = %+v", invocation)
		}
	}
}
