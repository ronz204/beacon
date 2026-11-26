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

func TestStatusHoldsEverythingTheCappedBriefOmitted(t *testing.T) {
	project, root := newProject(t)
	writeFile(t, root, configFile, strings.Replace(
		readFile(t, root, configFile),
		"brief: 60",
		"brief: 3",
		1,
	))

	brief, briefFindings := project.Brief(context.Background(), nil)
	status, statusFindings := project.Status(context.Background(), nil)

	if len(briefFindings) != 0 || len(statusFindings) != 0 {
		t.Fatalf("findings = %+v %+v", briefFindings, statusFindings)
	}
	if brief.Omitted == 0 || status.Omitted != 0 {
		t.Fatalf("omitted = %d and %d", brief.Omitted, status.Omitted)
	}
	if len(status.Lines) != len(brief.Lines)-1+brief.Omitted {
		t.Fatalf("lines = %d, brief = %+v", len(status.Lines), brief)
	}
	if len(status.Signals) == 0 {
		t.Fatal("status must carry the signals it was built from")
	}
}
