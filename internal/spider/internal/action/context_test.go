package action_test

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/action"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/thread"
	"github.com/ronz204/beacon/internal/spider/internal/driver/double"
	"github.com/ronz204/beacon/internal/spider/internal/driver/match"
)

var validRoot = filepath.Join("..", "driver", "record", "testdata", "valid")

const notRun = thread.SignalNotRunHere

func promiseKey(raw string) ledger.PromiseID {
	id, err := ledger.ParsePromiseID(raw)
	if err != nil {
		panic(err)
	}

	return id
}

const evidenceCommit = "9f3c2ab"

var testToday = mustDay("2026-11-24")

func mustDay(raw string) kernel.CalendarDate {
	day, err := kernel.ParseCalendarDate(raw)
	if err != nil {
		panic(err)
	}

	return day
}

func proven(outcome ledger.ProofOutcome) ledger.ProofResult {
	return ledger.ProofResult{Outcome: outcome, Commit: evidenceCommit}
}

func sourcesFor(root string, proofs double.Proofs) action.Sources {
	return action.Sources{
		Repositories: repositories(root),
		Proofs:       proofs,
		Changes:      double.Changes{evidenceCommit: nil},
		History:      double.History{},
		Matcher:      match.NewMatcher(),
		Today:        testToday,
	}
}

func buildContext(
	t *testing.T,
	proofs double.Proofs,
	paths ...string,
) action.ContextResult {
	t.Helper()

	result, failures := action.Context(
		context.Background(),
		sourcesFor(validRoot, proofs),
		paths,
	)
	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}

	return result
}

func hasSignal(
	signals []thread.Signal,
	kind thread.SignalKind,
	subject string,
) bool {
	id, err := kernel.ParseRecordID(subject)
	if err != nil {
		panic(err)
	}

	return slices.ContainsFunc(signals, func(signal thread.Signal) bool {
		return signal.Kind == kind && signal.Subject == id
	})
}

func TestContextGathersEverythingForAnOwnedFile(t *testing.T) {
	result := buildContext(t, nil, "src/auth/login.go")

	if len(result.Goals) != 1 {
		t.Fatalf("goals = %d", len(result.Goals))
	}
	if len(result.Capabilities) != 1 {
		t.Fatalf("capabilities = %d", len(result.Capabilities))
	}

	capability := result.Capabilities[0]
	if len(capability.Capability.Nongoals()) != 1 {
		t.Fatal("nongoals must travel with the capability")
	}
	if len(capability.Promises) != 2 {
		t.Fatalf("promises = %d", len(capability.Promises))
	}
	if len(result.Decisions) != 1 {
		t.Fatalf("decisions = %d", len(result.Decisions))
	}
	if len(result.Decisions[0].Rejected()) != 2 {
		t.Fatal("rejected alternatives must travel with the decision")
	}
	if len(result.Guides) != 1 {
		t.Fatalf("guides = %d", len(result.Guides))
	}
	if len(result.Questions) != 1 {
		t.Fatalf("questions = %d", len(result.Questions))
	}
}

func TestContextOfAnUnownedFileHasOnlyGoalAndAlwaysGuides(t *testing.T) {
	result := buildContext(t, nil, "docs/readme.md")

	if len(result.Goals) != 1 || len(result.Guides) != 1 {
		t.Fatalf("goals = %d, guides = %d",
			len(result.Goals),
			len(result.Guides),
		)
	}
	if len(result.Capabilities) != 0 || len(result.Decisions) != 0 {
		t.Fatalf("result = %+v", result)
	}
	if len(result.Questions) != 0 || len(result.Signals) != 0 {
		t.Fatalf("result = %+v", result)
	}
}

func TestContextReportsPromiseStates(t *testing.T) {
	proofs := double.Proofs{promiseKey("bc-c3d4"): proven(ledger.ProofFailed)}

	result := buildContext(t, proofs, "src/auth/login.go")

	states := map[string]ledger.VerificationStatus{}
	for _, promise := range result.Capabilities[0].Promises {
		states[promise.Promise.ID().String()] = promise.Verification.Status()
	}

	if states["bc-c3d4"] != ledger.VerificationViolated {
		t.Fatalf("states = %v", states)
	}
	if states["bc-c4a1"] != ledger.VerificationAttested {
		t.Fatalf("states = %v", states)
	}
}

func TestContextSignals(t *testing.T) {
	failed := double.Proofs{promiseKey("bc-c3d4"): proven(ledger.ProofFailed)}
	passed := double.Proofs{promiseKey("bc-c3d4"): proven(ledger.ProofPassed)}

	withFailure := buildContext(t, failed, "src/auth/login.go").Signals
	withPass := buildContext(t, passed, "src/auth/login.go").Signals
	withoutResult := buildContext(t, nil, "src/auth/login.go").Signals

	cases := []struct {
		name    string
		signals []thread.Signal
		kind    thread.SignalKind
		subject string
		want    bool
	}{
		{"red when failed", withFailure, thread.SignalRed, "bc-c3d4", true},
		{"no red when passed", withPass, thread.SignalRed, "bc-c3d4", false},
		{"not run here", withoutResult, notRun, "bc-c3d4", true},
		{"run here", withPass, notRun, "bc-c3d4", false},
		{"manual never pending", withoutResult, notRun, "bc-c4a1", false},
		{"blocked work", withPass, thread.SignalBlocked, "bc-91cd", true},
		{"tension to work", withPass, thread.SignalTension, "bc-c3d4", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasSignal(c.signals, c.kind, c.subject); got != c.want {
				t.Fatalf("got %v, want %v in %+v", got, c.want, c.signals)
			}
		})
	}
}

func TestSignalsReportUnscopedPromises(t *testing.T) {
	root := writeRecords(t, map[string]string{
		"capabilities/bc-7f3a-sample.yaml": "id: bc-7f3a\n" +
			"created: 2026-11-22\nauthor: \"human:jane\"\nname: Sample\n" +
			"intent: Sample\npromises:\n  - id: bc-c3d4\n    text: Sample\n" +
			"    check: {runner: test-go, target: x}\n    status: active\n",
		"capabilities/bc-7f3b-sample.yaml": "id: bc-7f3b\n" +
			"created: 2026-11-22\nauthor: \"human:jane\"\nname: Sample\n" +
			"intent: Sample\npaths: [src/**]\npromises:\n" +
			"  - id: bc-c3d5\n    text: Sample\n" +
			"    check: {runner: test-go, target: x}\n    status: active\n",
	})
	snapshot := action.Load(context.Background(), repositories(root))
	graph, failures := action.Validate(snapshot)
	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}

	verifications, failures := action.Verify(
		context.Background(),
		snapshot,
		double.Proofs{},
	)
	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}

	signals := action.Signals(snapshot, graph, verifications)

	if !hasSignal(signals, thread.SignalUnscoped, "bc-c3d4") {
		t.Fatalf("missing unscoped signal: %+v", signals)
	}
	if hasSignal(signals, thread.SignalUnscoped, "bc-c3d5") {
		t.Fatalf("scoped promise reported: %+v", signals)
	}
}

func TestContextAcceptsSeveralPaths(t *testing.T) {
	result := buildContext(t, nil, "docs/readme.md", "tests/auth/x_test.go")

	if len(result.Capabilities) != 1 {
		t.Fatalf("capabilities = %d", len(result.Capabilities))
	}
}
