package scan_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ronz204/beacon/source/spider/internal/driver/scan"
)

func build(t *testing.T, files map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for name, text := range files {
		target := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func runnerNames(report scan.Report) []string {
	names := []string{}
	for _, runner := range report.Runners {
		names = append(names, runner.Name)
	}

	return names
}

func TestScanFindsAGoModuleAndItsFolders(t *testing.T) {
	root := build(t, map[string]string{
		"go.mod":                    "module sample\n",
		"source/billing/billing.go": "package billing\n",
		"source/billing/tax.go":     "package billing\n",
		"source/auth/auth.go":       "package auth\n",
		"docs/guide.md":             "text\n",
		"node_modules/x/index.js":   "x\n",
	})

	report, err := scan.Scan(root)

	if err != nil || !slices.Equal(report.Stacks, []string{"go"}) {
		t.Fatalf("report = %+v, err = %v", report, err)
	}
	if !slices.Equal(runnerNames(report), []string{"test-go"}) {
		t.Errorf("runners = %+v", report.Runners)
	}
	want := []scan.Candidate{
		{Path: "source/billing", Files: 2},
		{Path: "source/auth", Files: 1},
	}
	if !slices.Equal(report.Candidates, want) {
		t.Errorf("candidates = %+v", report.Candidates)
	}
}

func TestScanReadsABunWorkspace(t *testing.T) {
	root := build(t, map[string]string{
		"bun.lock":                  "{}\n",
		"package.json":              `{"workspaces":["packages/*"]}`,
		"packages/api/package.json": "{}\n",
		"packages/api/index.ts":     "export {}\n",
		"packages/web/package.json": "{}\n",
		"packages/web/app.tsx":      "export {}\n",
	})

	report, err := scan.Scan(root)

	if err != nil || !slices.Equal(report.Stacks, []string{"bun"}) {
		t.Fatalf("report = %+v, err = %v", report, err)
	}
	if report.Runners[0].Run[0] != "bun" {
		t.Errorf("runners = %+v", report.Runners)
	}
	if len(report.Candidates) != 2 ||
		report.Candidates[0].Path != "packages/api" {
		t.Errorf("candidates = %+v", report.Candidates)
	}
}

func TestScanPicksTheNodeRunnerOfTheLockfile(t *testing.T) {
	root := build(t, map[string]string{
		"pnpm-lock.yaml":         "lock\n",
		"pnpm-workspace.yaml":    "packages:\n  - apps/*\n",
		"package.json":           "{}\n",
		"apps/site/package.json": "{}\n",
		"apps/site/main.js":      "x\n",
	})

	report, _ := scan.Scan(root)

	if !slices.Equal(report.Stacks, []string{"node"}) ||
		report.Runners[0].Run[0] != "pnpm" {
		t.Fatalf("report = %+v", report)
	}
	if len(report.Packages) != 2 {
		t.Errorf("packages = %+v", report.Packages)
	}
}

func TestScanDetectsPythonAndAMixedMonorepo(t *testing.T) {
	root := build(t, map[string]string{
		"api/go.mod":            "module api\n",
		"api/main.go":           "package main\n",
		"worker/pyproject.toml": "[project]\n",
		"worker/app/run.py":     "x = 1\n",
	})

	report, _ := scan.Scan(root)

	if !slices.Equal(report.Stacks, []string{"go", "python"}) {
		t.Fatalf("stacks = %v", report.Stacks)
	}
	if !slices.Equal(runnerNames(report), []string{"test-go", "test-python"}) {
		t.Errorf("runners = %v", runnerNames(report))
	}
}

func TestScanOfAnEmptyRepositoryHasNothing(t *testing.T) {
	report, err := scan.Scan(build(t, map[string]string{"notes.txt": "x\n"}))

	if err != nil || len(report.Stacks) != 0 || len(report.Runners) != 0 ||
		len(report.Candidates) != 0 {
		t.Fatalf("report = %+v, err = %v", report, err)
	}
}
