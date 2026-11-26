package claude_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/agents/claude"
)

const editPayload = `{
  "session_id": "sample-session",
  "cwd": "C:\\work\\sample",
  "hook_event_name": "PreToolUse",
  "tool_name": "Edit",
  "tool_input": {"file_path": "C:\\work\\sample\\src\\main.go"}
}`

func TestParseReadsTheSessionAndTheEditedFile(t *testing.T) {
	input, err := claude.Agent{}.ParseHook(
		agents.EventEdit,
		[]byte(editPayload),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := agents.HookInput{
		Session: "sample-session",
		Cwd:     `C:\work\sample`,
		Path:    `C:\work\sample\src\main.go`,
	}
	if input != want {
		t.Errorf("input = %+v, want %+v", input, want)
	}
}

func TestParseSkipsEditsOfOtherTools(t *testing.T) {
	cases := []string{
		`{"session_id":"s","tool_name":"Read","tool_input":{"file_path":"a"}}`,
		`{"session_id":"s","tool_name":"Edit","tool_input":{}}`,
		`{"session_id":"s","tool_name":"NotebookEdit","tool_input":{}}`,
	}

	for _, payload := range cases {
		input, err := claude.Agent{}.ParseHook(
			agents.EventEdit,
			[]byte(payload),
		)
		if err != nil || !input.Skip {
			t.Errorf("payload %s: input %+v, err %v", payload, input, err)
		}
	}
}

func TestParseAcceptsWriteAndStart(t *testing.T) {
	write := `{"session_id":"s","tool_name":"Write","tool_input":` +
		`{"file_path":"a.go"}}`
	input, err := claude.Agent{}.ParseHook(agents.EventEdit, []byte(write))
	if err != nil || input.Skip {
		t.Errorf("write: input %+v, err %v", input, err)
	}

	start := `{"session_id":"s","source":"compact"}`
	input, err = claude.Agent{}.ParseHook(agents.EventStart, []byte(start))
	if err != nil || input.Skip || input.Session != "s" {
		t.Errorf("start: input %+v, err %v", input, err)
	}
}

func TestParseRejectsBrokenInput(t *testing.T) {
	for _, payload := range []string{``, `nope`, `{}`, `{"cwd":"x"}`} {
		_, err := claude.Agent{}.ParseHook(agents.EventStart, []byte(payload))
		if err == nil {
			t.Errorf("payload %q should fail", payload)
		}
	}
}

type rendered struct {
	Specific struct {
		EventName string `json:"hookEventName"`
		Context   string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

func TestRenderWrapsTheTextForEachEvent(t *testing.T) {
	cases := map[agents.Event]string{
		agents.EventStart: "SessionStart",
		agents.EventEdit:  "PreToolUse",
	}

	for event, name := range cases {
		data, err := claude.Agent{}.RenderHook(event, "goal <a> & b")
		if err != nil {
			t.Fatal(err)
		}

		var output rendered
		if err := json.Unmarshal(data, &output); err != nil {
			t.Fatal(err)
		}

		if output.Specific.EventName != name {
			t.Errorf("%s: event = %q", event, output.Specific.EventName)
		}
		if output.Specific.Context != "goal <a> & b" {
			t.Errorf("%s: context = %q", event, output.Specific.Context)
		}
		if strings.Contains(string(data), `\u003c`) {
			t.Errorf("%s: HTML was escaped in %s", event, data)
		}
	}
}

func TestRenderNeverDecidesPermissions(t *testing.T) {
	data, err := claude.Agent{}.RenderHook(agents.EventEdit, "text")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(data), "permissionDecision") {
		t.Errorf("output decides permissions: %s", data)
	}
}

func TestRenderOfNothingPrintsNothing(t *testing.T) {
	data, err := claude.Agent{}.RenderHook(agents.EventStart, "")
	if err != nil || len(data) != 0 {
		t.Errorf("data %q, err %v", data, err)
	}
}
