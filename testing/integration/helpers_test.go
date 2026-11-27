package integration_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

const (
	capabilityPath = ".beacon/records/capabilities/bc-7f3a-auth-login.yaml"
	configPath     = ".beacon/config.yaml"
	sampleAgent    = "agent:sample-agent"
	lockoutText    = "After 5 failed attempts the account locks for 15 minutes."
)

type finding struct {
	Severity string         `json:"severity"`
	Code     string         `json:"code"`
	File     string         `json:"file"`
	Line     int            `json:"line"`
	Details  map[string]any `json:"details"`
}

type failure struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details struct {
		Findings []finding `json:"findings"`
		Changed  []string  `json:"changed"`
	} `json:"details"`
}

type signal struct {
	Kind   string `json:"kind"`
	Record string `json:"record"`
}

type envelope struct {
	OK      bool            `json:"ok"`
	Data    json.RawMessage `json:"data"`
	Signals []signal        `json:"signals"`
	Hint    *string         `json:"hint"`
	Error   *failure        `json:"error"`
}

func parseEnvelope(t *testing.T, text string) envelope {
	t.Helper()

	var parsed envelope
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid json %q: %v", text, err)
	}
	if parsed.Signals == nil {
		t.Fatalf("signals must be a list: %s", text)
	}

	return parsed
}

func decodeData(t *testing.T, text string, target any) {
	t.Helper()

	if err := json.Unmarshal(parseEnvelope(t, text).Data, target); err != nil {
		t.Fatal(err)
	}
}

func hintContains(hint *string, text string) bool {
	return hint != nil && strings.Contains(*hint, text)
}

func decodeError(t *testing.T, text string, target any) {
	t.Helper()

	var parsed struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(parsed.Error, target); err != nil {
		t.Fatal(err)
	}
}

func jsonOf(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func newBareRepository(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	utilities.Git(t, root, "init", "-q", "-b", "main")
	utilities.Git(t, root, "config", "user.name", "jane")

	return root
}

func present(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))

	return !errors.Is(err, os.ErrNotExist)
}
