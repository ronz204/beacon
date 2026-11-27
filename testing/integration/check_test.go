package integration_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

const (
	passingPromise = "bc-c3d4"
	failingPromise = "bc-c5b2"
	failingBlock   = `  - id: bc-c5b2
    text: Locked accounts show a neutral message.
    check: {runner: fake, target: fail}
    status: proposed
`
)

type checkData struct {
	Base     string   `json:"base"`
	Scope    string   `json:"scope"`
	Manual   []string `json:"manual"`
	Promises []struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		State   string `json:"state"`
		Drifted bool   `json:"drifted"`
	} `json:"promises"`
	Invocations []struct {
		Status string `json:"status"`
		Output string `json:"output"`
	} `json:"invocations"`
	Capabilities []string `json:"capabilities"`
}

func repositoryWithRunner(t *testing.T) string {
	t.Helper()

	root := utilities.NewRepository(t)
	config := filepath.Join(root, filepath.FromSlash(configPath))
	utilities.WriteFile(t, root, configPath, strings.Replace(
		utilities.ReadFile(t, config),
		"runners:\n",
		"runners:\n  fake:\n    run: ["+quoted(utilities.FakeRunner())+
			", \"{target}\"]\n",
		1,
	))

	capability := filepath.Join(root, filepath.FromSlash(capabilityPath))
	text := strings.Replace(
		utilities.ReadFile(t, capability),
		"check: {runner: test-node, pkg: api, target: auth/lockout}",
		"check: {runner: fake, target: pass}",
		1,
	)
	utilities.WriteFile(t, root, capabilityPath, text+failingBlock)

	if formatted := utilities.Beacon(t, root, "fmt"); formatted.Code != 0 {
		t.Fatalf("fmt failed: %s", formatted.Stderr)
	}

	utilities.Git(t, root, "add", "-A")
	utilities.Git(t, root, "commit", "-q", "-m", "add the fake runner")

	return root
}

func quoted(text string) string {
	return "\"" + text + "\""
}

func decodeCheck(t *testing.T, text string) checkData {
	t.Helper()

	var data checkData
	decodeData(t, text, &data)

	return data
}

func TestCheckRunsEveryPromiseAndBlocksOnARed(t *testing.T) {
	root := repositoryWithRunner(t)

	result := utilities.Beacon(t, root, "check", "bc-7f3a", "--json")

	data := decodeCheck(t, result.Stdout)
	envelope := parseEnvelope(t, result.Stdout)
	if result.Code != 2 || !envelope.OK {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}

	statuses := map[string]string{}
	for _, promise := range data.Promises {
		statuses[promise.ID] = promise.Status
	}
	if statuses[passingPromise] != "passed" {
		t.Fatalf("statuses = %v", statuses)
	}
	if statuses[failingPromise] != "failed" {
		t.Fatalf("statuses = %v", statuses)
	}
	if len(data.Manual) != 1 || data.Manual[0] != "bc-c4a1" {
		t.Fatalf("manual = %v", data.Manual)
	}
}

func TestCheckBlockedHintNamesThePromiseAndForbidsEditingIt(t *testing.T) {
	root := repositoryWithRunner(t)

	result := utilities.Beacon(t, root, "check", "bc-7f3a", "--json")

	envelope := parseEnvelope(t, result.Stdout)
	if !hintContains(envelope.Hint, failingPromise) {
		t.Fatalf("hint = %v", envelope.Hint)
	}
	if !hintContains(envelope.Hint, "beacon check bc-7f3a") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
	if !hintContains(envelope.Hint, "do not edit the promise") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
	if len(envelope.Signals) == 0 {
		t.Fatalf("the red must come as a signal: %s", result.Stdout)
	}
}

func TestCheckKeepsTheOutputOfAFailedInvocation(t *testing.T) {
	root := repositoryWithRunner(t)

	result := utilities.Beacon(t, root, "check", "bc-7f3a", "--json")

	var failedOutput string
	for _, invocation := range decodeCheck(t, result.Stdout).Invocations {
		if invocation.Status == "failed" {
			failedOutput = invocation.Output
		}
	}

	if !strings.Contains(failedOutput, "assertion failed") {
		t.Fatalf("output = %q", failedOutput)
	}
}

func TestCheckStreamsTheRunnerOutputToStderrInTextMode(t *testing.T) {
	root := repositoryWithRunner(t)

	result := utilities.Beacon(t, root, "check", "bc-7f3a")

	if !strings.Contains(result.Stderr, "assertion failed") {
		t.Fatalf("stderr = %q", result.Stderr)
	}
	if !strings.Contains(result.Stdout, "failed bc-c5b2") {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}

func makePassing(t *testing.T, root string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	utilities.WriteFile(t, root, capabilityPath, strings.Replace(
		utilities.ReadFile(t, path),
		"target: fail",
		"target: pass",
		1,
	))
	utilities.Git(t, root, "commit", "-q", "-am", "make it pass")
}

func TestCheckWithoutArgumentsOnlyRunsTheTouchedCapabilities(t *testing.T) {
	root := repositoryWithRunner(t)
	makePassing(t, root)

	idle := utilities.Beacon(t, root, "check", "--json")
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")
	busy := utilities.Beacon(t, root, "check", "--json")

	idleData := decodeCheck(t, idle.Stdout)
	busyData := decodeCheck(t, busy.Stdout)
	if idle.Code != 0 || len(idleData.Capabilities) != 0 {
		t.Fatalf("idle = %+v", idle)
	}
	if busy.Code != 0 || len(busyData.Capabilities) != 1 {
		t.Fatalf("busy = %+v", busy)
	}
	if busyData.Scope != "changed" || busyData.Base == "" {
		t.Fatalf("busy = %+v", busyData)
	}
}

func TestCheckRightAfterRunningShowsNoDrift(t *testing.T) {
	root := repositoryWithRunner(t)
	makePassing(t, root)
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")

	result := utilities.Beacon(t, root, "check", "bc-7f3a", "--json")

	for _, promise := range decodeCheck(t, result.Stdout).Promises {
		if promise.Drifted || promise.State != "holds" {
			t.Fatalf("promise = %+v", promise)
		}
	}

	utilities.WriteFile(
		t, root, "src/auth/login.go", "package auth // edited\n",
	)
	context := utilities.Beacon(
		t, root, "context", "src/auth/login.go", "--json",
	)
	if !strings.Contains(context.Stdout, `"drifted":true`) {
		t.Fatalf("editing after the check must drift: %s", context.Stdout)
	}
}

func TestCheckOfAnUnknownCapabilityHintsAtFind(t *testing.T) {
	root := repositoryWithRunner(t)

	result := utilities.Beacon(t, root, "check", "bc-ffff", "--json")

	envelope := parseEnvelope(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "capability_missing" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestCheckFailsBeforeRunningWhenARunnerIsMissing(t *testing.T) {
	root := repositoryWithRunner(t)
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	utilities.WriteFile(t, root, capabilityPath, strings.Replace(
		utilities.ReadFile(t, path),
		"runner: fake, target: pass",
		"runner: ghost, target: pass",
		1,
	))

	result := utilities.Beacon(t, root, "check", "bc-7f3a", "--json")

	envelope := parseEnvelope(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "runner_missing" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !hintContains(envelope.Hint, "ghost") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
}

func TestCheckWithAMissingProgramLeavesThePromiseUnverified(t *testing.T) {
	root := repositoryWithRunner(t)
	makePassing(t, root)
	config := filepath.Join(root, filepath.FromSlash(configPath))
	utilities.WriteFile(t, root, configPath, strings.Replace(
		utilities.ReadFile(t, config),
		"runners:\n",
		"runners:\n  absent:\n    run: [not-installed-program-xyz, "+
			"\"{target}\"]\n",
		1,
	))
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	utilities.WriteFile(t, root, capabilityPath, strings.Replace(
		utilities.ReadFile(t, path),
		"runner: fake, target: pass",
		"runner: absent, target: pass",
		1,
	))

	result := utilities.Beacon(t, root, "check", "bc-7f3a", "--json")

	envelope := parseEnvelope(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "runner_unavailable" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !hintContains(envelope.Hint, "not-installed-program-xyz") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
	if strings.Contains(result.Stdout, `"violated"`) {
		t.Fatalf("an unavailable runner is not a violation: %s", result.Stdout)
	}
}

func TestCheckNeedsABaseWhenTheRepositoryHasNone(t *testing.T) {
	root := repositoryWithRunner(t)
	utilities.Git(t, root, "branch", "-m", "trunk")

	missing := utilities.Beacon(t, root, "check", "--json")
	forced := utilities.Beacon(t, root, "check", "--base", "trunk", "--json")
	option := utilities.Beacon(t, root, "check", "--base=-x", "--json")

	var failureEnvelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	err := json.Unmarshal([]byte(missing.Stdout), &failureEnvelope)
	if err != nil {
		t.Fatal(err)
	}

	if missing.Code != 1 || failureEnvelope.Error.Code != "base_missing" {
		t.Fatalf("missing = %+v", missing)
	}
	if forced.Code != 0 {
		t.Fatalf("forced = %+v", forced)
	}
	if option.Code != 1 {
		t.Fatalf("option = %+v", option)
	}
}
