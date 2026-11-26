package e2e_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/utils"
)

type hookReply struct {
	Specific struct {
		EventName string `json:"hookEventName"`
		Context   string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

func hookPayload(t *testing.T, fields map[string]any) string {
	t.Helper()

	data, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func editPayload(t *testing.T, root, tool, file string) string {
	t.Helper()

	return hookPayload(t, map[string]any{
		"session_id": "sample-session",
		"cwd":        root,
		"tool_name":  tool,
		"tool_input": map[string]any{"file_path": file},
	})
}

func runHook(t *testing.T, root, event, stdin string) utils.Result {
	t.Helper()

	return utils.BeaconWith(
		t,
		root,
		utils.Options{Stdin: stdin},
		"hook", event, "--agent", "claude-code",
	)
}

func TestHookStartInjectsTheBrief(t *testing.T) {
	root := utils.NewRepository(t)
	stdin := hookPayload(t, map[string]any{
		"session_id": "sample-session",
		"cwd":        root,
		"source":     "compact",
	})

	result := runHook(t, root, "start", stdin)

	var reply hookReply
	if err := json.Unmarshal([]byte(result.Stdout), &reply); err != nil {
		t.Fatalf("stdout %q: %v", result.Stdout, err)
	}

	if result.Code != 0 || reply.Specific.EventName != "SessionStart" {
		t.Errorf("code %d, reply %+v", result.Code, reply)
	}
	if strings.TrimSpace(reply.Specific.Context) == "" {
		t.Error("the brief was empty")
	}
}

func TestHookEditInjectsOnlyTheFirstTimePerFile(t *testing.T) {
	root := utils.NewRepository(t)
	file := filepath.Join(root, "src", "auth", "login.go")
	stdin := editPayload(t, root, "Edit", file)

	first := runHook(t, root, "edit", stdin)
	second := runHook(t, root, "edit", stdin)

	var reply hookReply
	if err := json.Unmarshal([]byte(first.Stdout), &reply); err != nil {
		t.Fatalf("stdout %q: %v", first.Stdout, err)
	}

	if first.Code != 0 || reply.Specific.EventName != "PreToolUse" {
		t.Errorf("first: code %d, reply %+v", first.Code, reply)
	}
	if strings.Contains(first.Stdout, "permissionDecision") {
		t.Errorf("the hook decided permissions: %s", first.Stdout)
	}
	if second.Code != 0 || strings.TrimSpace(second.Stdout) != "" {
		t.Errorf("second: %+v", second)
	}
}

func TestHookEditIgnoresWhatIsNotAnEditInsideTheRepository(t *testing.T) {
	root := utils.NewRepository(t)
	outside := filepath.Join(t.TempDir(), "other.go")
	inside := filepath.Join(root, "src", "auth", "login.go")

	cases := []string{
		editPayload(t, root, "Read", inside),
		editPayload(t, root, "Edit", outside),
		editPayload(t, root, "Edit", ""),
	}

	for _, stdin := range cases {
		result := runHook(t, root, "edit", stdin)

		if result.Code != 0 || strings.TrimSpace(result.Stdout) != "" {
			t.Errorf("stdin %s: %+v", stdin, result)
		}
	}
}

func TestHookStaysSilentOutsideABeaconProject(t *testing.T) {
	root := newBareRepository(t)
	stdin := hookPayload(t, map[string]any{
		"session_id": "sample-session",
		"cwd":        root,
	})

	result := runHook(t, root, "start", stdin)

	if result.Code != 0 || result.Stdout != "" || result.Stderr != "" {
		t.Errorf("result = %+v", result)
	}
}

func TestHookNeverFailsTheSession(t *testing.T) {
	root := utils.NewRepository(t)
	good := hookPayload(t, map[string]any{
		"session_id": "sample-session",
		"cwd":        root,
	})

	broken := utils.NewRepository(t)
	utils.WriteFile(t, broken, ".beacon/config.yaml", "schema: [")
	brokenState := hookPayload(t, map[string]any{
		"session_id": "sample-session",
		"cwd":        broken,
	})

	cases := map[string]utils.Result{
		"invalid json":  runHook(t, root, "start", "nope"),
		"no session":    runHook(t, root, "start", `{"cwd":"x"}`),
		"bad event":     runHook(t, root, "stop", good),
		"invalid state": runHook(t, broken, "start", brokenState),
		"unknown agent": utils.BeaconWith(
			t, root, utils.Options{Stdin: good}, "hook", "start",
			"--agent", "sample-agent",
		),
	}

	for name, result := range cases {
		if result.Code != 0 || result.Stdout != "" {
			t.Errorf("%s: code %d, stdout %q", name, result.Code, result.Stdout)
		}
		if !strings.Contains(result.Stderr, "beacon hook:") {
			t.Errorf("%s: stderr %q", name, result.Stderr)
		}
	}
}
