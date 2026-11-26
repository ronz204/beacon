package anchor_test

import (
	"testing"

	"github.com/ronz204/beacon/source/anchor"
	"github.com/ronz204/beacon/source/kernel"
)

func codeOf(err error) string {
	return kernel.DescribeFailure(err).Code
}

func TestParseHookRejectsUnknownAgentsAndEvents(t *testing.T) {
	payload := []byte(`{"session_id":"s"}`)

	_, err := anchor.ParseHook("sample-agent", "start", payload)
	if codeOf(err) != "hook_agent_unknown" {
		t.Errorf("agent: err = %v", err)
	}

	_, err = anchor.ParseHook("claude-code", "stop", payload)
	if codeOf(err) != "hook_event_unknown" {
		t.Errorf("event: err = %v", err)
	}
}

func TestParseHookWrapsUnreadableInput(t *testing.T) {
	_, err := anchor.ParseHook("claude-code", "start", []byte("nope"))

	if codeOf(err) != "hook_input_invalid" {
		t.Errorf("err = %v", err)
	}
}

func TestHookRoundTripForClaudeCode(t *testing.T) {
	payload := []byte(`{"session_id":"s","cwd":"/tmp"}`)

	input, err := anchor.ParseHook("claude-code", "start", payload)
	if err != nil || input.Session != "s" {
		t.Fatalf("input %+v, err %v", input, err)
	}

	data, err := anchor.RenderHook("claude-code", "start", "brief")
	if err != nil || len(data) == 0 {
		t.Errorf("data %q, err %v", data, err)
	}
}
