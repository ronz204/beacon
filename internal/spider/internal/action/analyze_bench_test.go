package action_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/action"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/driver/commit"
	"github.com/ronz204/beacon/internal/spider/internal/driver/double"
	"github.com/ronz204/beacon/internal/spider/internal/driver/match"
)

const (
	benchmarkModules       = 500
	benchmarkCommits       = 50
	benchmarkFilesPerTouch = 20
)

func benchGit(b *testing.B, root string, args ...string) string {
	b.Helper()

	command := exec.Command("git", args...)
	command.Dir = root
	command.Env = append(
		os.Environ(),
		"GIT_AUTHOR_DATE=2026-11-23T12:00:00",
		"GIT_COMMITTER_DATE=2026-11-23T12:00:00",
	)

	output, err := command.CombinedOutput()
	if err != nil {
		b.Fatalf("git %v: %v\n%s", args, err, output)
	}

	return strings.TrimSpace(string(output))
}

func benchWrite(b *testing.B, root, relative, text string) {
	b.Helper()

	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		b.Fatal(err)
	}
}

func benchmarkCapabilityAt(index int) string {
	text := benchmarkCapability(index)

	return strings.Replace(
		text,
		fmt.Sprintf("paths: [src/module%d/**]", index),
		fmt.Sprintf(
			"paths: [src/module%d/**, tests/module%d/**]",
			index,
			index,
		),
		1,
	)
}

func analyzeRepository(b *testing.B) (string, double.Proofs) {
	b.Helper()

	root := b.TempDir()
	benchGit(b, root, "init", "-q")
	benchGit(b, root, "config", "user.name", "jane")
	benchGit(b, root, "config", "user.email", "jane@example.com")
	benchGit(b, root, "config", "commit.gpgsign", "false")

	for index := range benchmarkModules {
		name := fmt.Sprintf(
			".beacon/records/capabilities/bc-%04x-sample.yaml",
			index,
		)
		benchWrite(b, root, name, benchmarkCapabilityAt(index))
		benchWrite(b, root, fmt.Sprintf("src/module%d/a.go", index), "a")
		benchWrite(b, root, fmt.Sprintf("src/module%d/b.go", index), "b")
	}
	benchGit(b, root, "add", "-A")
	benchGit(b, root, "commit", "-q", "-m", "base")
	base := benchGit(b, root, "rev-parse", "HEAD")

	for change := range benchmarkCommits {
		for file := range benchmarkFilesPerTouch {
			module := (change*benchmarkFilesPerTouch + file) % benchmarkModules
			path := fmt.Sprintf("src/module%d/a.go", module)
			benchWrite(b, root, path, fmt.Sprintf("change %d", change))
		}
		benchGit(b, root, "commit", "-q", "-am", "change")
	}

	benchGit(b, root, "gc", "-q")

	for module := range 30 {
		path := fmt.Sprintf("src/module%d/b.go", module)
		benchWrite(b, root, path, "local edit")
	}

	proofs := double.Proofs{}
	for index := range benchmarkModules {
		for promise := range benchmarkPromises {
			raw := fmt.Sprintf(
				"bc-%05x",
				index*benchmarkPromises+promise+0x10000,
			)
			proofs[promiseKey(raw)] = ledger.ProofResult{
				Outcome: ledger.ProofPassed,
				Commit:  base,
			}
		}
	}

	return root, proofs
}

func BenchmarkAnalyzeWithRealGit(b *testing.B) {
	root, proofs := analyzeRepository(b)
	state := filepath.Join(root, ".beacon")
	repository := commit.NewRepository(root)
	sources := action.Sources{
		Repositories: repositories(state),
		Proofs:       proofs,
		Changes:      repository,
		History:      repository,
		Matcher:      match.NewMatcher(),
		Today:        testToday,
	}
	ctx := context.Background()

	for b.Loop() {
		analysis, failures := action.Analyze(ctx, sources)
		if len(failures) != 0 {
			b.Fatalf("failures = %v", failures)
		}

		drifted := 0
		for _, verification := range analysis.Verifications {
			if verification.Drifted() {
				drifted++
			}
		}
		if drifted == 0 {
			b.Fatal("expected drift from the commits")
		}
	}
}

func BenchmarkContextWithRealGit(b *testing.B) {
	root, proofs := analyzeRepository(b)
	state := filepath.Join(root, ".beacon")
	repository := commit.NewRepository(root)
	sources := action.Sources{
		Repositories: repositories(state),
		Proofs:       proofs,
		Changes:      repository,
		History:      repository,
		Matcher:      match.NewMatcher(),
		Today:        testToday,
	}
	ctx := context.Background()

	for b.Loop() {
		result, failures := action.Context(
			ctx,
			sources,
			[]string{"src/module7/a.go"},
		)
		if len(failures) != 0 || len(result.Capabilities) != 1 {
			b.Fatalf("failures = %v, result = %+v", failures, result)
		}
	}
}

func BenchmarkHookWithRealGit(b *testing.B) {
	root, proofs := analyzeRepository(b)
	state := filepath.Join(root, ".beacon")
	repository := commit.NewRepository(root)
	sources := action.Sources{
		Repositories: repositories(state),
		Proofs:       proofs,
		Changes:      repository,
		History:      repository,
		Matcher:      match.NewMatcher(),
		Today:        testToday,
	}
	ctx := context.Background()
	session, err := kernel.ParseSessionID("bench-session")
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, failures := action.HookContext(ctx, sources, action.HookInput{
			Session: session,
			Paths:   []string{"src/module7/a.go"},
			Limit:   20,
			Log:     double.Injections{},
		})
		if len(failures) != 0 {
			b.Fatalf("failures = %v", failures)
		}
	}
}
