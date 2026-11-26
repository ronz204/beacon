package e2e_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

type summary struct {
	Lines []struct {
		Section string `json:"section"`
		Text    string `json:"text"`
	} `json:"lines"`
	Omitted int `json:"omitted"`
}

func decodeSummary(t *testing.T, text string) summary {
	t.Helper()

	var data summary
	decodeData(t, text, &data)

	return data
}

func texts(data summary) []string {
	lines := make([]string, 0, len(data.Lines))
	for _, line := range data.Lines {
		lines = append(lines, line.Text)
	}

	return lines
}

func limitBrief(t *testing.T, root, limit string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(configPath))
	utils.WriteFile(t, root, configPath, strings.Replace(
		utils.ReadFile(t, path),
		"brief: 60",
		"brief: "+limit,
		1,
	))
}

func TestBriefStartsWithTheGoalAndCarriesSignals(t *testing.T) {
	root := utils.NewRepository(t)

	result := utils.Beacon(t, root, "brief", "--json")

	data := decodeSummary(t, result.Stdout)
	envelope := parseRead(t, result.Stdout)
	if result.Code != 0 || len(data.Lines) == 0 {
		t.Fatalf("result = %+v", result)
	}
	if !strings.HasPrefix(data.Lines[0].Text, "Goal bc-0001") {
		t.Fatalf("lines = %v", texts(data))
	}
	if len(envelope.Signals) == 0 {
		t.Fatalf("the brief must carry its signals: %s", result.Stdout)
	}
}

func TestBriefListsTheYesesThatArePending(t *testing.T) {
	root := utils.NewRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")

	result := utils.Beacon(t, root, "brief", "--json")

	data := decodeSummary(t, result.Stdout)
	envelope := parseRead(t, result.Stdout)
	joined := strings.Join(texts(data), "\n")
	if !strings.Contains(joined, "Needs approval: promise bc-c3d4") {
		t.Fatalf("lines = %v", texts(data))
	}
	if !strings.Contains(joined, "After 10 failed attempts") {
		t.Fatalf("the pending line must quote the new text: %s", joined)
	}

	pending := 0
	for _, signal := range envelope.Signals {
		if signal.Kind == "pending_yes" && signal.Record == "bc-c3d4" {
			pending++
		}
	}
	if pending != 1 {
		t.Fatalf("signals = %+v", envelope.Signals)
	}
}

func TestCappedBriefNamesStatusAndStatusHoldsTheRest(t *testing.T) {
	root := utils.NewRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")
	limitBrief(t, root, "4")

	brief := decodeSummary(
		t,
		utils.Beacon(t, root, "brief", "--json").Stdout,
	)
	status := decodeSummary(
		t,
		utils.Beacon(t, root, "status", "--json").Stdout,
	)

	last := brief.Lines[len(brief.Lines)-1]
	if brief.Omitted == 0 || !strings.Contains(last.Text, "beacon status") {
		t.Fatalf("brief = %+v", brief)
	}
	if status.Omitted != 0 {
		t.Fatalf("status must not cap: %+v", status)
	}
	if len(status.Lines) != len(brief.Lines)-1+brief.Omitted {
		t.Fatalf("status = %v, brief = %v", texts(status), texts(brief))
	}

	statusLines := strings.Join(texts(status), "\n")
	for _, line := range brief.Lines[:len(brief.Lines)-1] {
		if !strings.Contains(statusLines, line.Text) {
			t.Fatalf("status lost %q", line.Text)
		}
	}
}

func TestStatusSummarisesTheCountsAndNeverBlocks(t *testing.T) {
	root := repositoryWithRunner(t)
	failing := utils.Beacon(t, root, "check", "bc-7f3a", "--json")
	editByHand(t, root, "After 5 failed", "After 10 failed")

	result := utils.Beacon(t, root, "status")

	if failing.Code != 2 {
		t.Fatalf("setup = %+v", failing)
	}
	if result.Code != 0 {
		t.Fatalf("status only informs: %+v", result)
	}
	first := strings.SplitN(result.Stdout, "\n", 2)[0]
	if !strings.Contains(first, "1 red") {
		t.Fatalf("first line = %q", first)
	}
	if !strings.Contains(first, "1 pending yes") {
		t.Fatalf("first line = %q", first)
	}
	if !strings.Contains(result.Stdout, "Needs approval") {
		t.Fatalf("stdout = %s", result.Stdout)
	}
}

func TestStatusOfAnInvalidProjectFailsWithAHint(t *testing.T) {
	root := utils.NewRepository(t)
	breakGoal(t, root)

	result := utils.Beacon(t, root, "status", "--json")

	var parsed struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Stdout), &parsed); err != nil {
		t.Fatal(err)
	}

	if result.Code != 1 || parsed.Error.Code == "" {
		t.Fatalf("result = %+v", result)
	}
}
