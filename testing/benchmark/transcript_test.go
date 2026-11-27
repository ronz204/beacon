package benchmark_test

import (
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/benchmark"
)

func TestTheIdealTranscriptFollowsTheMethod(t *testing.T) {
	session := replay(t, "ideal")

	if !session.StartContext || !session.AgentBrief || !session.BriefAtStart() {
		t.Errorf("brief: %+v", session)
	}
	if len(session.Edits) != 1 || !session.Edits[0].AgentContext ||
		!session.Edits[0].HookContext {
		t.Errorf("edits = %+v", session.Edits)
	}
	if session.Tampering != 0 || session.BlockedByBeacon != 0 {
		t.Errorf("integrity: %+v", session)
	}
	if session.Commands["check"] != 1 || session.Injections != 2 {
		t.Errorf("commands %v, injections %d",
			session.Commands, session.Injections)
	}
	if session.Turns != 6 || session.CostUSD == 0 {
		t.Errorf("result fields: %+v", session)
	}
}

func TestTheIgnoringTranscriptIsToldApart(t *testing.T) {
	session := replay(t, "ignoring")

	if session.BriefAtStart() || session.StartContext {
		t.Errorf("it never asked for the brief: %+v", session)
	}
	if session.EditsWithContext() != 0 || len(session.Edits) != 2 {
		t.Errorf("edits = %+v", session.Edits)
	}
	if session.Tampering != 1 {
		t.Errorf("tampering = %d, want 1", session.Tampering)
	}
	if session.BlockedByBeacon != 1 || session.BlockedByPermission != 1 {
		t.Errorf("blocks: beacon %d, permission %d",
			session.BlockedByBeacon, session.BlockedByPermission)
	}
}

func TestARealClaudeCodeTranscriptParses(t *testing.T) {
	session := replay(t, "real-claude-code")

	if !session.StartContext || !session.BriefAtStart() {
		t.Errorf("the start hook was not seen: %+v", session)
	}
	if len(session.Edits) != 1 || !session.Edits[0].HookContext ||
		!session.Edits[0].AgentContext {
		t.Errorf("edits = %+v", session.Edits)
	}
	if session.Commands["context"] == 0 || session.Commands["check"] == 0 {
		t.Errorf("commands = %v", session.Commands)
	}
	if session.BlockedByPermission == 0 {
		t.Errorf("the permission wait was not seen: %+v", session)
	}
	if session.Turns == 0 || session.CostUSD == 0 {
		t.Errorf("result fields: %+v", session)
	}
}

func TestParseRejectsBrokenLines(t *testing.T) {
	_, err := benchmark.ParseTranscript(strings.NewReader("not json\n"))

	if err == nil {
		t.Error("expected an error")
	}
}

func TestParseSkipsLinesWhoseMessageIsPlainText(t *testing.T) {
	lines := `{"type":"user","message":"plain text, not a block list"}` + "\n" +
		`{"type":"result","num_turns":3,"duration_ms":10,"total_cost_usd":0.5}`

	session, err := benchmark.ParseTranscript(strings.NewReader(lines))
	if err != nil {
		t.Fatal(err)
	}

	if session.Turns != 3 || len(session.Edits) != 0 {
		t.Errorf("session = %+v", session)
	}
}
