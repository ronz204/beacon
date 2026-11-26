package spider_test

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func gitOutput(t *testing.T, root string, args ...string) string {
	t.Helper()

	command := exec.Command("git", args...)
	command.Dir = root

	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(output))
}

func TestCheckTargetsListEveryPromiseWithItsCheck(t *testing.T) {
	project, _ := newProject(t)

	targets, findings := project.CheckTargets(context.Background())

	if len(findings) != 0 || len(targets) != 1 {
		t.Fatalf("targets = %+v, findings = %+v", targets, findings)
	}

	promise := targets[0].Promises[0]
	if promise.ID != "bc-c3d4" || promise.Runner != "test-node" {
		t.Fatalf("promise = %+v", promise)
	}
	if promise.Params["target"] != "auth/lockout" {
		t.Fatalf("params = %v", promise.Params)
	}
	if targets[0].Promises[1].Manual == "" {
		t.Fatalf("the manual promise lost its steps: %+v", targets[0])
	}
}

func TestTouchedCapabilitiesFollowTheChangedFiles(t *testing.T) {
	project, root := newProject(t)
	head := gitOutput(t, root, "rev-parse", "HEAD")

	none, err := project.TouchedCapabilities(context.Background(), head)
	if err != nil {
		t.Fatal(err)
	}

	writeSource(t, root, "src/auth/login.go", "package auth")
	some, err := project.TouchedCapabilities(context.Background(), head)
	if err != nil {
		t.Fatal(err)
	}

	if len(none) != 0 || !slices.Equal(some, []string{"bc-7f3a"}) {
		t.Fatalf("none = %v, some = %v", none, some)
	}
}
