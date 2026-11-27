package integration_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

const manualID = "bc-c4a1"

type attestData struct {
	Pending []struct {
		ID      string `json:"id"`
		State   string `json:"state"`
		Drifted bool   `json:"drifted"`
		Manual  string `json:"manual"`
	} `json:"pending"`
	Attested []string `json:"attested"`
	Commit   string   `json:"commit"`
}

func featureBranchWithEdit(t *testing.T) string {
	t.Helper()

	root := utilities.NewRepository(t)
	utilities.Git(t, root, "checkout", "-q", "-b", "feature")
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")
	utilities.Git(t, root, "add", ".")
	utilities.Git(t, root, "commit", "-q", "-m", "edit login")

	return root
}

func runAttest(
	t *testing.T,
	root, actor string,
	args ...string,
) (utilities.Result, attestData) {
	t.Helper()

	arguments := append([]string{"attest", "--json"}, args...)
	result := utilities.BeaconWith(
		t, root, utilities.Options{Actor: actor}, arguments...,
	)

	var data attestData
	if result.Code == 0 {
		decodeData(t, result.Stdout, &data)
	}

	return result, data
}

func TestAttestListsNothingWhenNoCodeChanged(t *testing.T) {
	root := utilities.NewRepository(t)

	result, data := runAttest(t, root, "")

	if result.Code != 0 || len(data.Pending) != 0 {
		t.Fatalf("code %d, data %+v", result.Code, data)
	}
}

func TestAttestListsTheManualPromiseWhoseCodeChanged(t *testing.T) {
	root := featureBranchWithEdit(t)

	result, data := runAttest(t, root, "")

	if result.Code != 0 || len(data.Pending) != 1 {
		t.Fatalf("code %d, stdout %s", result.Code, result.Stdout)
	}
	if data.Pending[0].ID != manualID || !data.Pending[0].Drifted {
		t.Errorf("pending = %+v", data.Pending[0])
	}
	if data.Pending[0].Manual == "" {
		t.Error("the steps to review are missing")
	}

	text := utilities.Beacon(t, root, "attest")
	if !strings.Contains(text.Stdout, "ask the person") {
		t.Errorf("the hint does not tell the agent to ask:\n%s", text.Stdout)
	}
}

func TestAttestRecordsThePersonAndTheEvidence(t *testing.T) {
	root := featureBranchWithEdit(t)

	result, data := runAttest(
		t, root, "human:sam", manualID, "--evidence", "docs/login.png",
	)

	if result.Code != 0 || len(data.Attested) != 1 || data.Commit == "" {
		t.Fatalf("code %d, stdout %s", result.Code, result.Stdout)
	}

	text := utilities.ReadFile(t, filepath.Join(root, capabilityPath))
	for _, want := range []string{
		`by: "human:sam"`, data.Commit, "docs/login.png",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}

	_, after := runAttest(t, root, "")
	if len(after.Pending) != 0 {
		t.Errorf("still pending after the attestation: %+v", after.Pending)
	}
}

func TestAttestByAnAgentKeepsThePersonAndTheAgent(t *testing.T) {
	root := featureBranchWithEdit(t)

	result, _ := runAttest(t, root, sampleAgent, manualID)

	text := utilities.ReadFile(t, filepath.Join(root, capabilityPath))
	if result.Code != 0 || !strings.Contains(text, `by: "human:jane"`) ||
		!strings.Contains(text, `via: "agent:sample-agent"`) {
		t.Fatalf("code %d, text:\n%s", result.Code, text)
	}
}

func TestAttestRefusesUncommittedCode(t *testing.T) {
	root := utilities.NewRepository(t)
	utilities.WriteFile(t, root, "src/auth/login.go", "package auth\n")

	result, _ := runAttest(t, root, "human:sam", manualID)

	refused := strings.Contains(result.Stdout, "attest_uncommitted")
	if result.Code != 1 || !refused {
		t.Fatalf("code %d, stdout %s", result.Code, result.Stdout)
	}
	if !strings.Contains(result.Stdout, "commit the code") {
		t.Errorf("the hint does not say what to do: %s", result.Stdout)
	}
}

func TestAttestRefusesAnAutomaticPromise(t *testing.T) {
	root := utilities.NewRepository(t)

	result, _ := runAttest(t, root, "human:sam", "bc-c3d4")

	refused := strings.Contains(result.Stdout, "attest_not_manual")
	if result.Code != 1 || !refused {
		t.Fatalf("code %d, stdout %s", result.Code, result.Stdout)
	}
}
