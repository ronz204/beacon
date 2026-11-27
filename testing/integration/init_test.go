package integration_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

type initData struct {
	DryRun bool `json:"dryRun"`
	Scan   struct {
		Stacks  []string `json:"stacks"`
		Runners []struct {
			Name string   `json:"name"`
			Run  []string `json:"run"`
		} `json:"runners"`
		Candidates []struct {
			Path string `json:"path"`
		} `json:"candidates"`
	} `json:"scan"`
	Records []struct {
		Type string `json:"type"`
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"records"`
	Omitted  int    `json:"omitted"`
	Proposal string `json:"proposal"`
}

func goRepository(t *testing.T, folders ...string) string {
	t.Helper()

	root := newBareRepository(t)
	utilities.WriteFile(t, root, "go.mod", "module sample\n")
	for _, folder := range folders {
		utilities.WriteFile(t, root, folder+"/code.go", "package code\n")
	}

	return root
}

func runInit(
	t *testing.T,
	root string,
	args ...string,
) (utilities.Result, initData) {
	t.Helper()

	arguments := append([]string{"init", "--json"}, args...)
	result := utilities.BeaconWith(
		t,
		root,
		utilities.Options{Actor: sampleAgent},
		arguments...,
	)

	var data initData
	if result.Code == 0 {
		decodeData(t, result.Stdout, &data)
	}

	return result, data
}

func TestInitLeavesTheRepositoryReadyToAsk(t *testing.T) {
	root := goRepository(t, "source/billing", "source/auth")

	result, data := runInit(t, root)

	if result.Code != 0 || len(data.Records) != 5 {
		t.Fatalf("code %d, data %+v: %s", result.Code, data, result.Stderr)
	}
	if !present(root, configPath) {
		t.Fatal("config.yaml was not written")
	}

	doctor := utilities.Beacon(t, root, "doctor")
	if doctor.Code != 0 {
		t.Fatalf("doctor: %s%s", doctor.Stdout, doctor.Stderr)
	}

	brief := utilities.Beacon(t, root, "brief")
	if !strings.Contains(brief.Stdout, "goal of this project") {
		t.Errorf("brief = %s", brief.Stdout)
	}
}

func TestInitProposesRunnersThatEnterWithAYes(t *testing.T) {
	root := goRepository(t, "source/billing")
	_, data := runInit(t, root)

	if len(data.Scan.Runners) != 1 || data.Scan.Runners[0].Name != "test-go" {
		t.Fatalf("runners = %+v", data.Scan.Runners)
	}

	file := utilities.WriteFile(t, t.TempDir(), "proposal.json", data.Proposal)
	blocked := utilities.BeaconWith(
		t,
		root,
		utilities.Options{Actor: sampleAgent},
		"record", "--json", "--file", file,
	)
	approved := utilities.BeaconWith(
		t,
		root,
		utilities.Options{Actor: sampleAgent},
		"record", "--json", "--approve", "--file", file,
	)

	text := utilities.ReadFile(t, root+"/"+configPath)
	if blocked.Code != 2 || approved.Code != 0 ||
		!strings.Contains(text, "test-go") {
		t.Fatalf("blocked %+v, approved %+v\n%s", blocked, approved, text)
	}
}

func TestInitDryRunChangesNothing(t *testing.T) {
	root := goRepository(t, "source/billing")

	result, data := runInit(t, root, "--dry-run")

	if result.Code != 0 || !data.DryRun || len(data.Records) == 0 {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
	if present(root, ".beacon") {
		t.Error("dry run wrote .beacon")
	}
}

func TestInitRefusesARepositoryThatAlreadyHasBeacon(t *testing.T) {
	root := goRepository(t, "source/billing")
	runInit(t, root)

	result, _ := runInit(t, root)

	if result.Code != 1 || !strings.Contains(result.Stdout, "project_exists") {
		t.Fatalf("code %d, stdout %s", result.Code, result.Stdout)
	}
}

func TestInitWithoutAKnownStackProposesNothing(t *testing.T) {
	root := newBareRepository(t)
	utilities.WriteFile(t, root, "notes.txt", "hello\n")

	result, data := runInit(t, root)

	envelope := parseEnvelope(t, result.Stdout)
	if result.Code != 0 || len(data.Scan.Stacks) != 0 || data.Proposal != "" {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
	if !hintContains(envelope.Hint, "no known stack") {
		t.Errorf("hint = %v", envelope.Hint)
	}
}

func TestInitKeepsTenCapabilitiesAndAsksAboutTheRest(t *testing.T) {
	var folders []string
	for index := range 12 {
		folders = append(folders, fmt.Sprintf("source/area%02d", index))
	}
	root := goRepository(t, folders...)

	_, data := runInit(t, root)

	capabilities := 0
	for _, record := range data.Records {
		if record.Type == "capability" {
			capabilities++
		}
	}
	if capabilities != 10 || data.Omitted != 2 {
		t.Fatalf("capabilities %d, omitted %d", capabilities, data.Omitted)
	}
	if last := data.Records[len(data.Records)-1]; last.Type != "question" ||
		!strings.Contains(last.Text, "2 more folders") {
		t.Errorf("last record = %+v", last)
	}
}
