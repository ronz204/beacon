package action_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
	"github.com/ronz204/beacon/source/spider/internal/driver/double"
)

func passedProofs() double.Proofs {
	return double.Proofs{promiseKey("bc-c3d4"): proven(ledger.ProofPassed)}
}

func analyzeWith(
	t *testing.T,
	sources action.Sources,
) (action.Analysis, []error) {
	t.Helper()

	return action.Analyze(context.Background(), sources)
}

func requireClean(t *testing.T, failures []error) {
	t.Helper()

	if len(failures) != 0 {
		t.Fatalf("failures = %v", failures)
	}
}

func driftedPromises(analysis action.Analysis) []string {
	var drifted []string

	for id, verification := range analysis.Verifications {
		if verification.Drifted() {
			drifted = append(drifted, id.String())
		}
	}

	slices.Sort(drifted)

	return drifted
}

func repeatCommits(count int, raw string, files ...string) []memory.CodeCommit {
	commits := make([]memory.CodeCommit, 0, count)

	for range count {
		commits = append(commits, memory.CodeCommit{
			Day:   mustDay(raw),
			Files: files,
		})
	}

	return commits
}

func TestDriftAppearsWhenScopedCodeChanged(t *testing.T) {
	sources := sourcesFor(validRoot, passedProofs())
	sources.Changes = double.Changes{evidenceCommit: {"src/auth/login.go"}}

	analysis, failures := analyzeWith(t, sources)

	requireClean(t, failures)
	got := driftedPromises(analysis)
	if !slices.Equal(got, []string{"bc-c3d4", "bc-c4a1"}) {
		t.Fatalf("drifted = %v", got)
	}
	if !hasSignal(analysis.Signals, thread.SignalDrift, "bc-c3d4") {
		t.Fatalf("signals = %+v", analysis.Signals)
	}
}

func TestDriftIgnoresChangesOutsideThePaths(t *testing.T) {
	sources := sourcesFor(validRoot, passedProofs())
	sources.Changes = double.Changes{evidenceCommit: {"docs/readme.md"}}

	analysis, failures := analyzeWith(t, sources)

	requireClean(t, failures)
	if got := driftedPromises(analysis); len(got) != 0 {
		t.Fatalf("drifted = %v", got)
	}
}

func TestDriftNeverAppliesToViolatedPromises(t *testing.T) {
	proofs := double.Proofs{promiseKey("bc-c3d4"): proven(ledger.ProofFailed)}
	sources := sourcesFor(validRoot, proofs)
	sources.Changes = double.Changes{evidenceCommit: {"src/auth/login.go"}}

	analysis, failures := analyzeWith(t, sources)

	requireClean(t, failures)
	got := driftedPromises(analysis)
	if !slices.Equal(got, []string{"bc-c4a1"}) {
		t.Fatalf("drifted = %v", got)
	}
}

func TestDriftIsAssumedWhenTheCommitIsUnknown(t *testing.T) {
	sources := sourcesFor(validRoot, passedProofs())
	sources.Changes = double.Changes{}

	analysis, failures := analyzeWith(t, sources)

	requireClean(t, failures)
	if got := driftedPromises(analysis); len(got) != 2 {
		t.Fatalf("drifted = %v", got)
	}
}

func TestStaleWhenReviewIsOverdue(t *testing.T) {
	sources := sourcesFor(validRoot, nil)
	sources.Today = mustDay("2027-06-01")

	analysis, failures := analyzeWith(t, sources)

	requireClean(t, failures)
	if len(analysis.Stale) != 1 || analysis.Stale[0].ID.String() != "bc-a1b2" {
		t.Fatalf("stale = %+v", analysis.Stale)
	}
	if !analysis.Stale[0].Overdue {
		t.Fatal("the decision should be overdue")
	}
	if !hasSignal(analysis.Signals, thread.SignalStale, "bc-a1b2") {
		t.Fatalf("signals = %+v", analysis.Signals)
	}
}

func TestStaleAfterTenCommitsOnItsPaths(t *testing.T) {
	sources := sourcesFor(validRoot, nil)
	ten := repeatCommits(10, "2026-11-23", "src/auth/login.go")
	nine := repeatCommits(9, "2026-11-23", "src/auth/login.go")
	early := repeatCommits(10, "2026-11-21", "src/auth/login.go")

	sources.History = double.History{Commits: ten}
	analysis, _ := analyzeWith(t, sources)
	if len(analysis.Stale) != 2 {
		t.Fatalf("with ten commits stale = %+v", analysis.Stale)
	}

	sources.History = double.History{Commits: nine}
	analysis, _ = analyzeWith(t, sources)
	if len(analysis.Stale) != 0 {
		t.Fatalf("with nine commits stale = %+v", analysis.Stale)
	}

	sources.History = double.History{Commits: slices.Concat(nine, early)}
	analysis, _ = analyzeWith(t, sources)
	if len(analysis.Stale) != 0 {
		t.Fatalf("commits before the review counted: %+v", analysis.Stale)
	}
}

func TestStaleGuideWithoutReviewUsesItsLastCommit(t *testing.T) {
	guide := "---\ntitle: Sample\nkind: runbook\n---\nBody\n"
	root := writeRecords(t, map[string]string{
		"guides/bc-0a9b-sample.md": guide,
		"guides/bc-0a9c-sample.md": guide,
	})
	sources := sourcesFor(root, nil)
	sources.Review = map[string]memory.ReviewSpan{
		"runbook": must6Months(t),
	}
	sources.History = double.History{
		Guides: map[memory.GuideID]kernel.CalendarDate{
			guideKey(t, "bc-0a9b"): mustDay("2025-01-01"),
		},
	}

	analysis, failures := analyzeWith(t, sources)

	requireClean(t, failures)
	if len(analysis.Stale) != 1 || analysis.Stale[0].ID.String() != "bc-0a9b" {
		t.Fatalf("stale = %+v", analysis.Stale)
	}
}

func TestStaleIgnoresSupersededDecisions(t *testing.T) {
	decision := "id: bc-a1b2\ncreated: 2020-01-01\n" +
		"author: \"human:jane\"\nwhat: Sample\nwhy: Sample\n" +
		"status: superseded\nreview: 1d\n"
	root := writeRecords(t, map[string]string{
		"decisions/bc-a1b2-sample.yaml": decision,
	})

	analysis, failures := analyzeWith(t, sourcesFor(root, nil))

	requireClean(t, failures)
	if len(analysis.Stale) != 0 {
		t.Fatalf("stale = %+v", analysis.Stale)
	}
}

func TestBriefListsDriftAndStaleRecords(t *testing.T) {
	sources := sourcesFor(validRoot, passedProofs())
	sources.Changes = double.Changes{evidenceCommit: {"src/auth/login.go"}}
	sources.Today = mustDay("2027-06-01")

	result, failures := action.Brief(context.Background(), sources, 60, nil)

	requireClean(t, failures)
	text := strings.Join(briefTexts(result), "\n")
	expected := []string{
		"Drift: promise bc-c3d4: After 5 failed",
		"Drift: promise bc-c4a1: The sign-in screen",
		"Stale: decision bc-a1b2: Sessions with 15 min JWTs",
	}
	for _, line := range expected {
		if !strings.Contains(text, line) {
			t.Fatalf("missing %q in:\n%s", line, text)
		}
	}
}

func must6Months(t *testing.T) memory.ReviewSpan {
	t.Helper()

	span, err := memory.ParseReviewSpan("6mo")
	if err != nil {
		t.Fatal(err)
	}

	return span
}

func guideKey(t *testing.T, raw string) memory.GuideID {
	t.Helper()

	id, err := memory.ParseGuideID(raw)
	if err != nil {
		t.Fatal(err)
	}

	return id
}
