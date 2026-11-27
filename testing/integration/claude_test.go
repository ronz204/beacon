package integration_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utilities"
)

const existingSettings = `{
  "model": "sample-model",
  "permissions": {
    "allow": [
      "Bash(go test *)"
    ],
    "deny": [
      "Read(**/.env)"
    ]
  }
}
`

type claudeSettings struct {
	Model       string `json:"model"`
	Permissions struct {
		Allow []string `json:"allow"`
		Deny  []string `json:"deny"`
	} `json:"permissions"`
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
	} `json:"hooks"`
}

func readSettings(t *testing.T, root string) claudeSettings {
	t.Helper()

	settings := filepath.Join(root, ".claude", "settings.json")
	text := utilities.ReadFile(t, settings)

	var parsed claudeSettings
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("invalid settings %q: %v", text, err)
	}

	return parsed
}

func TestInstallKeepsWhatClaudeSettingsAlreadyHad(t *testing.T) {
	root := newBareRepository(t)
	utilities.WriteFile(t, root, ".claude/settings.json", existingSettings)

	runInstall(t, root)
	parsed := readSettings(t, root)

	if parsed.Model != "sample-model" || len(parsed.Permissions.Deny) != 1 {
		t.Errorf("settings lost content: %+v", parsed)
	}

	allow := strings.Join(parsed.Permissions.Allow, ",")
	if allow != "Bash(go test *),Bash(beacon *),PowerShell(beacon *)" {
		t.Errorf("allow = %q", allow)
	}
	starts := len(parsed.Hooks["SessionStart"])
	if starts != 2 || len(parsed.Hooks["PreToolUse"]) != 1 {
		t.Errorf("hooks = %+v", parsed.Hooks)
	}
}

func TestUninstallGivesTheSettingsBackByteForByte(t *testing.T) {
	root := newBareRepository(t)
	utilities.WriteFile(t, root, ".claude/settings.json", existingSettings)

	runInstall(t, root)
	runInstall(t, root, "--uninstall")

	path := filepath.Join(root, ".claude", "settings.json")
	if utilities.ReadFile(t, path) != existingSettings {
		t.Errorf("settings changed:\n%s", utilities.ReadFile(t, path))
	}
}

func TestInstallImportsTheGuideIntoAnExistingClaudeFile(t *testing.T) {
	root := newBareRepository(t)
	utilities.WriteFile(t, root, "CLAUDE.md", "# My rules\n")

	runInstall(t, root)

	text := utilities.ReadFile(t, filepath.Join(root, "CLAUDE.md"))
	if !strings.HasPrefix(text, "# My rules\n") ||
		!strings.Contains(text, "@AGENTS.md") {
		t.Errorf("CLAUDE.md = %q", text)
	}

	runInstall(t, root, "--uninstall")

	restored := utilities.ReadFile(t, filepath.Join(root, "CLAUDE.md"))
	if restored != "# My rules\n" {
		t.Error("CLAUDE.md was not restored")
	}
}

func TestInstallSkipsWhatTheFlagsLeaveOut(t *testing.T) {
	cases := []struct {
		args  []string
		start int
		edit  int
	}{
		{nil, 2, 1},
		{[]string{"--skip", "reinject"}, 1, 0},
		{[]string{"--skip", "hooks"}, 0, 0},
	}

	for _, c := range cases {
		root := newBareRepository(t)
		runInstall(t, root, c.args...)

		parsed := readSettings(t, root)
		if len(parsed.Hooks["SessionStart"]) != c.start ||
			len(parsed.Hooks["PreToolUse"]) != c.edit {
			t.Errorf("args %v: hooks = %+v", c.args, parsed.Hooks)
		}
		if len(parsed.Permissions.Allow) != 2 {
			t.Errorf("args %v: allow = %v", c.args, parsed.Permissions.Allow)
		}
	}
}

func TestInstallReportsTheMomentClaudeCodeCannotCover(t *testing.T) {
	root := newBareRepository(t)

	result := utilities.Beacon(t, root, "install", "--json", "--dry-run")

	var data struct {
		Agents []struct {
			Name    string   `json:"name"`
			Missing []string `json:"missing"`
		} `json:"agents"`
	}
	decodeData(t, result.Stdout, &data)

	if len(data.Agents) != 1 || data.Agents[0].Name != "claude-code" {
		t.Fatalf("agents = %+v", data.Agents)
	}
	if strings.Join(data.Agents[0].Missing, ",") != "stop" {
		t.Errorf("missing = %v", data.Agents[0].Missing)
	}
}
