package integration_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

type verifyData struct {
	Risk struct {
		Level   string   `json:"level"`
		Reasons []string `json:"reasons"`
	} `json:"risk"`
	Check  checkData `json:"check"`
	Review *struct {
		Promises []struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"promises"`
		Changes []struct {
			ID    string `json:"id"`
			Level int    `json:"level"`
		} `json:"changes"`
		Rules []string `json:"rules"`
	} `json:"review"`
}

func runVerify(
	t *testing.T,
	root string,
	args ...string,
) (utilities.Result, verifyData) {
	t.Helper()

	arguments := append([]string{"verify", "--json"}, args...)
	result := utilities.Beacon(t, root, arguments...)

	var data verifyData
	if result.Code == 0 || result.Code == 2 {
		decodeData(t, result.Stdout, &data)
	}

	return result, data
}

func passingRepository(t *testing.T) string {
	t.Helper()

	root := repositoryWithRunner(t)
	makePassing(t, root)

	return root
}

func TestVerifyIsLowRiskWhenNothingWasTouched(t *testing.T) {
	root := passingRepository(t)

	result, data := runVerify(t, root)

	if result.Code != 0 || data.Risk.Level != "low" || data.Review != nil {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
}

func TestVerifyRatesAWeakeningEditHighAndPreparesTheReview(t *testing.T) {
	root := passingRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")

	result, data := runVerify(t, root)

	if result.Code != 0 || data.Risk.Level != "high" {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
	if !slices.Contains(data.Risk.Reasons, "weakening_change") {
		t.Errorf("reasons = %v", data.Risk.Reasons)
	}
	if data.Review == nil || len(data.Review.Changes) == 0 ||
		len(data.Review.Rules) == 0 {
		t.Fatalf("the review packet is incomplete: %+v", data.Review)
	}
	if data.Review.Changes[0].Level != 3 {
		t.Errorf("changes = %+v", data.Review.Changes)
	}

	envelope := parseEnvelope(t, result.Stdout)
	if !hintContains(envelope.Hint, "clean context") {
		t.Errorf("hint = %v", envelope.Hint)
	}
}

func TestVerifyBlocksOnARedPromiseInATouchedZone(t *testing.T) {
	root := repositoryWithRunner(t)
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")

	result, data := runVerify(t, root)

	if result.Code != 2 || data.Risk.Level != "high" {
		t.Fatalf("code %d, stdout %s", result.Code, result.Stdout)
	}
	if !slices.Contains(data.Risk.Reasons, "red_promise") {
		t.Errorf("reasons = %v", data.Risk.Reasons)
	}
	if data.Review == nil || len(data.Review.Promises) == 0 {
		t.Errorf("the review lacks the touched promises: %+v", data.Review)
	}
}

func TestVerifyTheAgentCanRaiseTheRiskButNotLowerIt(t *testing.T) {
	root := passingRepository(t)

	_, raised := runVerify(t, root, "--risk", "high")
	if raised.Risk.Level != "high" ||
		!slices.Contains(raised.Risk.Reasons, "raised_on_request") {
		t.Errorf("raised = %+v", raised.Risk)
	}

	result := utilities.Beacon(t, root, "verify", "--json", "--risk", "low")
	rejected := strings.Contains(result.Stdout, "risk_invalid")
	if result.Code != 1 || !rejected {
		t.Errorf("code %d, stdout %s", result.Code, result.Stdout)
	}
}

func TestDiffShowsTheRisk(t *testing.T) {
	root := passingRepository(t)
	editByHand(t, root, "After 5 failed", "After 10 failed")

	text := utilities.Beacon(t, root, "diff")
	if !strings.Contains(text.Stdout, "risk: high (weakening_change") {
		t.Errorf("text diff:\n%s", text.Stdout)
	}

	var data struct {
		Risk struct {
			Level string `json:"level"`
		} `json:"risk"`
	}
	json := utilities.Beacon(t, root, "diff", "--json")
	decodeData(t, json.Stdout, &data)
	if data.Risk.Level != "high" {
		t.Errorf("risk = %+v", data.Risk)
	}
}
