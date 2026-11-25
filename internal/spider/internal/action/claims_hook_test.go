package action_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/action"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/internal/spider/internal/driver/double"
	"github.com/ronz204/beacon/internal/spider/internal/driver/record"
)

func localClaim(t *testing.T, task, branch, worktree string) agenda.TaskClaim {
	t.Helper()

	id, err := agenda.ParseTaskID(task)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := agenda.ParseTaskBranch(branch)
	if err != nil {
		t.Fatal(err)
	}

	return agenda.TaskClaim{
		Task:     id,
		Branch:   parsed,
		Worktree: worktree,
		Day:      testToday,
	}
}

func branchOf(t *testing.T, raw string) agenda.TaskBranch {
	t.Helper()

	branch, err := agenda.ParseTaskBranch(raw)
	if err != nil {
		t.Fatal(err)
	}

	return branch
}

func TestReadyRespectsLocalClaims(t *testing.T) {
	root := writeRecords(t, map[string]string{
		workFile:              workText("bc-91cd"),
		taskFile("bc-91cd.1"): openTask("bc-91cd.1"),
		taskFile("bc-91cd.2"): openTask("bc-91cd.2"),
	})
	claims := double.NewClaims(
		localClaim(t, "bc-91cd.1", "feat/sample-a", "/work/a"),
	)

	ready := func(branch string) []string {
		result, failures := action.Ready(
			context.Background(),
			repositories(root),
			claims,
			branchOf(t, branch),
		)
		requireClean(t, failures)
		return readyTaskIDs(result)
	}

	if got := ready("feat/sample-a"); !slices.Equal(got, []string{
		"bc-91cd.1",
		"bc-91cd.2",
	}) {
		t.Fatalf("own branch ready = %v", got)
	}
	if got := ready("feat/sample-b"); !slices.Equal(got, []string{
		"bc-91cd.2",
	}) {
		t.Fatalf("other branch ready = %v", got)
	}
}

func TestPruneClaimsRemovesGoneBranchesAndWorktrees(t *testing.T) {
	claims := double.NewClaims(
		localClaim(t, "bc-91cd.1", "feat/sample-a", "/work/a"),
		localClaim(t, "bc-91cd.2", "feat/gone", "/work/a"),
		localClaim(t, "bc-91cd.3", "feat/sample-a", "/work/gone"),
	)
	workspace := double.Workspace{
		Branches:  []agenda.TaskBranch{branchOf(t, "feat/sample-a")},
		Worktrees: []string{"/work/a"},
	}

	removed, err := action.PruneClaims(context.Background(), claims, workspace)
	if err != nil {
		t.Fatal(err)
	}

	if len(removed) != 2 {
		t.Fatalf("removed = %+v", removed)
	}

	left, _ := claims.ListClaims(context.Background())
	if len(left) != 1 || left[0].Task.String() != "bc-91cd.1" {
		t.Fatalf("left = %+v", left)
	}
}

func TestDoctorReportsRemovedClaims(t *testing.T) {
	findings := action.Doctor(context.Background(), action.DoctorInput{
		Repositories: repositories(validRoot),
		Formatter:    record.NewStore(validRoot),
		Claims: double.NewClaims(
			localClaim(t, "bc-91cd.1", "feat/gone", "/work/a"),
		),
		Workspace:  double.Workspace{Worktrees: []string{"/work/a"}},
		GuideLimit: 150,
	})

	if len(findings) != 1 || findings[0].Code != action.CodeClaimRemoved {
		t.Fatalf("findings = %+v", findings)
	}
	if findings[0].Severity != action.FindingWarning {
		t.Fatalf("severity = %q", findings[0].Severity)
	}
}

func hookOf(
	t *testing.T,
	log double.Injections,
	limit int,
	paths ...string,
) action.HookResult {
	t.Helper()

	sessionID, err := kernel.ParseSessionID("sample-session")
	if err != nil {
		t.Fatal(err)
	}

	result, failures := action.HookContext(
		context.Background(),
		sourcesFor(validRoot, nil),
		action.HookInput{
			Session: sessionID,
			Paths:   paths,
			Limit:   limit,
			Log:     log,
		},
	)
	requireClean(t, failures)

	return result
}

func TestHookInjectsEachRecordOncePerSession(t *testing.T) {
	log := double.Injections{}

	first := hookOf(t, log, 60, "src/auth/login.go")
	text := strings.Join(first.Lines, "\n")
	expected := []string{
		"Capability bc-7f3a: Inicio de sesión",
		"Not doing: No incluye inicio de sesión con redes sociales.",
		"Promise bc-c3d4 (unverified): Tras 5 intentos",
		"Decision bc-a1b2: Sesiones con JWT",
		"Rejected: Sesiones en servidor",
		"Guide bc-0a9b (architecture): Arquitectura general",
		"Open question bc-e5f6:",
		"Signal not_run_here: bc-c3d4",
	}
	for _, line := range expected {
		if !strings.Contains(text, line) {
			t.Fatalf("missing %q in:\n%s", line, text)
		}
	}

	second := hookOf(t, log, 60, "src/auth/login.go")
	if len(second.Lines) != 0 {
		t.Fatalf("repeated lines: %q", second.Lines)
	}
}

func TestHookRespectsTheLimit(t *testing.T) {
	log := double.Injections{}

	first := hookOf(t, log, 4, "src/auth/login.go")

	if len(first.Lines) > 4 {
		t.Fatalf("lines = %d: %q", len(first.Lines), first.Lines)
	}
	if first.Omitted == 0 {
		t.Fatal("something should have been omitted")
	}

	last := first.Lines[len(first.Lines)-1]
	if !strings.HasSuffix(last, "run beacon context src/auth/login.go") {
		t.Fatalf("last line = %q", last)
	}

	second := hookOf(t, log, 60, "src/auth/login.go")
	if len(second.Lines) == 0 {
		t.Fatal("omitted items must arrive on the next call")
	}
}

func TestHookDeliversAnItemLargerThanTheLimit(t *testing.T) {
	log := double.Injections{}

	first := hookOf(t, log, 1, "src/auth/login.go")
	if len(first.Lines) != 1 {
		t.Fatalf("lines = %q", first.Lines)
	}

	for range 20 {
		hookOf(t, log, 1, "src/auth/login.go")
	}

	last := hookOf(t, log, 1, "src/auth/login.go")
	if len(last.Lines) != 0 {
		t.Fatalf("the hook never settles: %q", last.Lines)
	}
}

func TestHookNeverQueriesGit(t *testing.T) {
	sessionID, err := kernel.ParseSessionID("sample-session")
	if err != nil {
		t.Fatal(err)
	}

	sources := sourcesFor(validRoot, passedProofs())
	sources.Changes = double.Changes{}

	result, failures := action.HookContext(
		context.Background(),
		sources,
		action.HookInput{
			Session: sessionID,
			Paths:   []string{"src/auth/login.go"},
			Limit:   60,
			Log:     double.Injections{},
		},
	)

	requireClean(t, failures)
	if strings.Contains(strings.Join(result.Lines, "\n"), "drift") {
		t.Fatalf("the hook reported drift: %q", result.Lines)
	}
}

func TestAnalyzeWithoutGitSkipsDriftAndFreshness(t *testing.T) {
	sources := sourcesFor(validRoot, passedProofs())
	sources.Changes = nil
	sources.History = nil
	sources.Today = mustDay("2030-01-01")

	analysis, failures := action.Analyze(context.Background(), sources)

	requireClean(t, failures)
	if len(driftedPromises(analysis)) != 0 || len(analysis.Stale) != 0 {
		t.Fatalf("analysis = %+v", analysis)
	}
}
