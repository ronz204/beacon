package integration_test

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

func TestGetPrintsTheRecordInTheEnvelope(t *testing.T) {
	root := utilities.NewRepository(t)

	result := utilities.Beacon(t, root, "get", "bc-7f3a", "--json")

	var data struct {
		Type     string `json:"type"`
		Archived bool   `json:"archived"`
	}
	envelope := parseEnvelope(t, result.Stdout)
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatal(err)
	}

	if result.Code != 0 || !envelope.OK || data.Type != "capability" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestGetOfAnUnknownIDHintsAtFind(t *testing.T) {
	root := utilities.NewRepository(t)

	result := utilities.Beacon(t, root, "get", "bc-ffff", "--json")

	envelope := parseEnvelope(t, result.Stdout)
	if result.Code != 1 || envelope.Error.Code != "record_missing" {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if !hintContains(envelope.Hint, "beacon find") {
		t.Fatalf("hint = %v", envelope.Hint)
	}
}

func TestFindListsMatchingRecords(t *testing.T) {
	root := utilities.NewRepository(t)

	result := utilities.Beacon(t, root, "find", "sign-in")

	if result.Code != 0 || !strings.Contains(result.Stdout, "bc-7f3a") {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestContextReturnsWhatAppliesWithSignals(t *testing.T) {
	root := utilities.NewRepository(t)

	result := utilities.Beacon(
		t, root, "context", "src/auth/login.go", "--json",
	)

	envelope := parseEnvelope(t, result.Stdout)
	if result.Code != 0 || !strings.Contains(string(envelope.Data), "bc-7f3a") {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	if len(envelope.Signals) == 0 {
		t.Fatalf("signals are empty: %s", result.Stdout)
	}
}

func TestContextOfAFileWithoutOwnerOnlyHasTheGoalAndGuides(t *testing.T) {
	root := utilities.NewRepository(t)

	result := utilities.Beacon(t, root, "context", "docs/notes.md", "--json")

	var data struct {
		Goals        []json.RawMessage `json:"goals"`
		Capabilities []json.RawMessage `json:"capabilities"`
		Decisions    []json.RawMessage `json:"decisions"`
	}
	decodeData(t, result.Stdout, &data)

	if len(data.Goals) != 1 || len(data.Capabilities) != 0 {
		t.Fatalf("stdout = %s", result.Stdout)
	}
	if len(data.Decisions) != 0 {
		t.Fatalf("stdout = %s", result.Stdout)
	}
}

func TestHookContextNeedsASession(t *testing.T) {
	root := utilities.NewRepository(t)

	result := utilities.Beacon(
		t, root, "context", "--hook", "src/auth/login.go",
	)

	if result.Code != 1 || !strings.Contains(result.Stderr, "--session") {
		t.Fatalf("code = %d, stderr = %s", result.Code, result.Stderr)
	}
}

func TestHookContextOnlySendsWhatIsNewInTheSession(t *testing.T) {
	root := utilities.NewRepository(t)
	args := []string{
		"context", "--hook", "--session", "sample-session", "src/auth/login.go",
	}

	first := utilities.Beacon(t, root, args...)
	second := utilities.Beacon(t, root, args...)

	if first.Code != 0 || strings.TrimSpace(first.Stdout) == "" {
		t.Fatalf("first = %+v", first)
	}
	if second.Code != 0 || strings.TrimSpace(second.Stdout) != "" {
		t.Fatalf("second = %+v", second)
	}
}

func TestReadyReportsTheCurrentBranch(t *testing.T) {
	root := utilities.NewRepository(t)

	result := utilities.Beacon(t, root, "ready", "--json")

	var data struct {
		Branch struct {
			Name     string `json:"name"`
			Detached bool   `json:"detached"`
		} `json:"branch"`
	}
	decodeData(t, result.Stdout, &data)

	if result.Code != 0 || data.Branch.Name != "main" || data.Branch.Detached {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestReadyWarnsOnADetachedHead(t *testing.T) {
	root := utilities.NewRepository(t)
	utilities.Git(t, root, "checkout", "-q", "--detach")

	result := utilities.Beacon(t, root, "ready")

	if result.Code != 0 || !strings.Contains(result.Stdout, "detached") {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
}

func TestContextSurvivesAnAttestedCommitThatNoLongerExists(t *testing.T) {
	root := utilities.NewRepository(t)
	path := filepath.Join(root, filepath.FromSlash(capabilityPath))
	text := utilities.ReadFile(t, path)
	attested := regexp.MustCompile(`commit: [0-9a-f]{7}`)
	utilities.WriteFile(
		t,
		root,
		capabilityPath,
		attested.ReplaceAllString(text, "commit: 9f3c2ab"),
	)

	result := utilities.Beacon(
		t, root, "context", "src/auth/login.go", "--json",
	)

	var data struct {
		Capabilities []struct {
			Promises []struct {
				ID      string `json:"id"`
				Drifted bool   `json:"drifted"`
			} `json:"promises"`
		} `json:"capabilities"`
	}
	decodeData(t, result.Stdout, &data)

	if result.Code != 0 {
		t.Fatalf("code = %d, stdout = %s", result.Code, result.Stdout)
	}
	for _, promise := range data.Capabilities[0].Promises {
		if promise.ID == "bc-c4a1" && !promise.Drifted {
			t.Fatalf("the attestation must show drift: %s", result.Stdout)
		}
	}
}
