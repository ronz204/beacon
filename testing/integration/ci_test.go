package integration_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

type ciData struct {
	Revision string `json:"revision"`
	Scope    string `json:"scope"`
	Risk     struct {
		Level string `json:"level"`
	} `json:"risk"`
	Blocks []struct {
		Kind  string   `json:"kind"`
		Items []string `json:"items"`
	} `json:"blocks"`
	Warnings []string `json:"warnings"`
	Checked  []string `json:"checked"`
	Skipped  []string `json:"skipped"`
}

func runCI(
	t *testing.T,
	root string,
	args ...string,
) (utilities.Result, ciData) {
	t.Helper()

	arguments := append([]string{"snares", "ci", "--json"}, args...)
	arguments = append(arguments, "--base", "main")
	result := utilities.Beacon(t, root, arguments...)

	var data ciData
	if result.Code == 0 || result.Code == 2 {
		decodeData(t, result.Stdout, &data)
	}

	return result, data
}

func ciKinds(data ciData) []string {
	kinds := make([]string, 0, len(data.Blocks))
	for _, block := range data.Blocks {
		kinds = append(kinds, block.Kind)
	}

	return kinds
}

func openPullRequest(t *testing.T, root string) {
	t.Helper()

	utilities.Git(t, root, "checkout", "-q", "-b", "pr")
}

func commitTracked(t *testing.T, root, message string) {
	t.Helper()

	utilities.Git(t, root, "commit", "-q", "-am", message)
}

func TestCIPassesACleanPullRequest(t *testing.T) {
	root := passingRepository(t)
	openPullRequest(t, root)
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")
	utilities.Git(t, root, "add", "-A")
	commitTracked(t, root, "touch the code")

	result, data := runCI(t, root)

	if result.Code != 0 || len(data.Blocks) != 0 {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
	if data.Scope != "all" || len(data.Checked) != 1 {
		t.Errorf("data = %+v", data)
	}
}

func TestCIBlocksALevelThreeChangeWithoutAYes(t *testing.T) {
	root := passingRepository(t)
	openPullRequest(t, root)
	editByHand(t, root, "After 5 failed", "After 10 failed")
	commitTracked(t, root, "weaken the lockout")

	result, data := runCI(t, root)

	if result.Code != 2 || !slices.Contains(ciKinds(data), "unapproved_change") {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
	if data.Risk.Level != "high" {
		t.Errorf("risk = %q", data.Risk.Level)
	}
}

func TestCIBlocksARedPromiseOutsideTheTouchedCapabilities(t *testing.T) {
	root := repositoryWithRunner(t)
	openPullRequest(t, root)
	utilities.WriteFile(t, root, "README.md", "hello\n")
	utilities.Git(t, root, "add", "-A")
	commitTracked(t, root, "unrelated change")

	all, allData := runCI(t, root)
	touched, touchedData := runCI(t, root, "--scope", "touched")

	if all.Code != 2 || !slices.Contains(ciKinds(allData), "failed_check") {
		t.Fatalf("all: code %d, data %+v", all.Code, allData)
	}
	if touched.Code != 0 || len(touchedData.Checked) != 0 {
		t.Fatalf("touched: code %d, data %+v", touched.Code, touchedData)
	}
}

func useNewRunner(t *testing.T, root string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(configPath))
	utilities.WriteFile(t, root, configPath, strings.Replace(
		utilities.ReadFile(t, path),
		"runners:\n",
		"runners:\n  fake2:\n    run: ["+quoted(utilities.FakeRunner())+
			", \"{target}\"]\n",
		1,
	))
	editByHand(
		t,
		root,
		"runner: fake, target: pass",
		"runner: fake2, target: pass",
	)

	if formatted := utilities.Beacon(t, root, "fmt"); formatted.Code != 0 {
		t.Fatalf("fmt failed: %s", formatted.Stderr)
	}
}

func TestCIDoesNotRunRunnersThePullRequestIntroducesWithoutAYes(t *testing.T) {
	root := passingRepository(t)
	openPullRequest(t, root)
	useNewRunner(t, root)
	commitTracked(t, root, "add a runner")

	result, data := runCI(t, root)

	if result.Code != 2 || !slices.Contains(ciKinds(data), "runners_unapproved") {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
	if !slices.Contains(data.Skipped, passingPromise) {
		t.Errorf("skipped = %v", data.Skipped)
	}
}

func approvedConfig(t *testing.T) string {
	t.Helper()

	runner := map[string]any{
		"run": []string{utilities.FakeRunner(), "{target}"},
	}

	return jsonOf(t, map[string]any{
		"type": "config",
		"record": map[string]any{
			"schema": 1,
			"runners": map[string]any{
				"fake":      runner,
				"fake2":     runner,
				"test-node": map[string]any{"run": []string{"pnpm", "{target}"}},
				"visual": map[string]any{
					"run": []string{
						"pnpm", "exec", "playwright", "test", "--grep",
						"{target}",
					},
					"baseline": []string{"tests/**/*-snapshots/**"},
				},
			},
			"review": map[string]any{
				"convention":   "6mo",
				"architecture": "12mo",
				"product":      "12mo",
				"domain":       "12mo",
				"runbook":      "6mo",
				"decision":     "6mo",
			},
			"limits": map[string]any{"guide": 150, "brief": 60, "hook": 20},
			"agents": []string{"claude-code", "codex", "cursor"},
		},
	})
}

func TestCIUsesThePullRequestRunnersOnceTheyHaveAYes(t *testing.T) {
	root := passingRepository(t)
	openPullRequest(t, root)
	useNewRunner(t, root)
	config := record(t, root, sampleAgent, approvedConfig(t), "--approve")
	promise := record(
		t,
		root,
		sampleAgent,
		jsonOf(t, storedPayload(t, root, "bc-7f3a")),
		"--approve",
	)
	commitTracked(t, root, "add a runner with a yes")

	result, data := runCI(t, root)

	if config.Code != 0 || promise.Code != 0 {
		t.Fatalf("config = %+v, promise = %+v", config, promise)
	}
	if result.Code != 0 || len(data.Skipped) != 0 {
		t.Fatalf("code %d, data %+v, stdout %s", result.Code, data, result.Stdout)
	}
}

func TestCIWritesTheCommentForThePullRequest(t *testing.T) {
	root := passingRepository(t)
	openPullRequest(t, root)
	editByHand(t, root, "After 5 failed", "After 10 failed")
	commitTracked(t, root, "weaken the lockout")
	comment := filepath.Join(t.TempDir(), "comment.md")

	result := utilities.Beacon(
		t,
		root,
		"snares", "ci", "--base", "main", "--comment", comment,
	)

	text, err := os.ReadFile(comment)
	if err != nil {
		t.Fatalf("no comment: %v", err)
	}
	for _, want := range []string{"## Beacon", "Blocked", "Level 3", "Risk"} {
		if !strings.Contains(string(text), want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if result.Code != 2 {
		t.Errorf("code = %d", result.Code)
	}
}

func TestCIRejectsAnUnknownScope(t *testing.T) {
	root := passingRepository(t)

	result := utilities.Beacon(t, root, "snares", "ci", "--scope", "some")

	if result.Code != 1 || !strings.Contains(result.Stderr, "scope_invalid") {
		t.Fatalf("code %d, stderr %s", result.Code, result.Stderr)
	}
}
