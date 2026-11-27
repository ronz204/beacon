package bench_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/bench"
	"github.com/ronz204/beacon/testing/utils"
)

const relativeRecord = ".beacon/records/capabilities/bc-a1b1-discounts.yaml"

func TestMain(m *testing.M) {
	os.Exit(utils.RunSuite(m))
}

func beaconIn(t *testing.T) bench.Beacon {
	t.Helper()

	return func(directory string, args ...string) (string, error) {
		result := utils.Beacon(t, directory, args...)
		if result.Code != 0 {
			return result.Stdout, fmt.Errorf(
				"beacon %v exited %d: %s%s",
				args, result.Code, result.Stdout, result.Stderr,
			)
		}

		return result.Stdout, nil
	}
}

func task(t *testing.T, name string) bench.Task {
	t.Helper()

	loaded, err := bench.LoadTask(
		filepath.Join("testdata", "tasks", name, "task.json"),
	)
	if err != nil {
		t.Fatal(err)
	}

	return loaded
}

func replay(t *testing.T, name string) bench.Session {
	t.Helper()

	agent := bench.Replay{
		Transcript: filepath.Join("testdata", "transcripts", name+".jsonl"),
	}

	stream, err := agent.Run(context.Background(), "", bench.Task{})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()

	session, err := bench.ParseTranscript(stream)
	if err != nil {
		t.Fatal(err)
	}

	return session
}

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

func TestAReplayNeverCallsAnAgent(t *testing.T) {
	agent := bench.Replay{
		Transcript: filepath.Join("testdata", "transcripts", "missing.jsonl"),
	}

	stream, err := agent.Run(context.Background(), "", bench.Task{})
	if err == nil {
		_, _ = io.Copy(io.Discard, stream)
		t.Fatal("a missing transcript should fail")
	}
}

func TestParseRejectsBrokenLines(t *testing.T) {
	_, err := bench.ParseTranscript(strings.NewReader("not json\n"))

	if err == nil {
		t.Error("expected an error")
	}
}

func TestSummaryTellsTheTwoTranscriptsApartInEveryMetric(t *testing.T) {
	results := []bench.Result{
		{Condition: bench.ConditionFull, Session: replay(t, "ideal")},
		{Condition: bench.ConditionWithout, Session: replay(t, "ignoring")},
	}

	summaries := bench.Summarize(results)
	if len(summaries) != 2 {
		t.Fatalf("summaries = %+v", summaries)
	}

	without, full := summaries[0], summaries[1]

	if without.BriefAtStart.Hits != 0 || full.BriefAtStart.Hits != 1 {
		t.Errorf("brief: %+v vs %+v", without.BriefAtStart, full.BriefAtStart)
	}
	if without.ContextBeforeEdit.Hits != 0 || full.ContextBeforeEdit.Hits != 1 {
		t.Errorf("context: %+v vs %+v",
			without.ContextBeforeEdit, full.ContextBeforeEdit)
	}
	if without.TamperPerSession != 1 || full.TamperPerSession != 0 {
		t.Errorf("tamper: %v vs %v",
			without.TamperPerSession, full.TamperPerSession)
	}
	if without.BlocksPerSession != 2 || full.BlocksPerSession != 0 {
		t.Errorf("blocks: %v vs %v",
			without.BlocksPerSession, full.BlocksPerSession)
	}
	if without.MeanInjectedLines != 0 || full.MeanInjectedLines == 0 {
		t.Errorf("lines: %v vs %v",
			without.MeanInjectedLines, full.MeanInjectedLines)
	}
}

func TestRateIntervalsAreWiderWithFewerSessions(t *testing.T) {
	small := bench.Rate{Hits: 1, Total: 2}
	large := bench.Rate{Hits: 50, Total: 100}

	smallLow, smallHigh := small.Interval()
	largeLow, largeHigh := large.Interval()

	if smallHigh-smallLow <= largeHigh-largeLow {
		t.Errorf("small %v-%v, large %v-%v",
			smallLow, smallHigh, largeLow, largeHigh)
	}

	empty := bench.Rate{}
	if low, high := empty.Interval(); low != 0 || high != 1 {
		t.Errorf("empty interval = %v-%v", low, high)
	}
}

func TestFormatNamesEveryColumnAndTheCaveat(t *testing.T) {
	results := []bench.Result{
		{Condition: bench.ConditionFull, Session: replay(t, "ideal")},
	}

	text := bench.Format(bench.Summarize(results))

	for _, want := range []string{"brief@start", "ctx-before-edit", "Wilson"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
}

func TestConditionsInstallWhatTheProtocolSays(t *testing.T) {
	beacon := beaconIn(t)
	task := task(t, "add-total")

	cases := map[bench.Condition]struct {
		settings bool
		starts   int
		edits    int
	}{
		bench.ConditionWithout:    {false, 0, 0},
		bench.ConditionGuide:      {true, 0, 0},
		bench.ConditionFull:       {true, 2, 1},
		bench.ConditionNoReinject: {true, 1, 0},
	}

	for condition, want := range cases {
		directory := t.TempDir()
		err := bench.Prepare(directory, task, condition, beacon)
		if err != nil {
			t.Fatalf("%s: %v", condition, err)
		}

		path := filepath.Join(directory, ".claude", "settings.json")
		_, statErr := os.Stat(path)
		if (statErr == nil) != want.settings {
			t.Errorf("%s: settings present = %v", condition, statErr == nil)
		}
		if !want.settings {
			continue
		}

		text := utils.ReadFile(t, path)
		gotStarts := strings.Count(text, "beacon hook start")
		gotEdits := strings.Count(text, "beacon hook edit")
		if gotStarts != want.starts || gotEdits != want.edits {
			t.Errorf("%s: start hooks %d, edit hooks %d",
				condition, gotStarts, gotEdits)
		}
	}
}

func TestThePreparedRepositoryIsHealthy(t *testing.T) {
	beacon := beaconIn(t)
	directory := t.TempDir()

	err := bench.Prepare(
		directory, task(t, "add-total"), bench.ConditionFull, beacon,
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := beacon(directory, "doctor"); err != nil {
		t.Errorf("doctor: %v", err)
	}
}

func TestAnUntouchedRepositoryDoesNotResolveTheTask(t *testing.T) {
	beacon := beaconIn(t)

	for _, name := range []string{"add-total", "tempt-weaken"} {
		directory := t.TempDir()
		chosen := task(t, name)

		err := bench.Prepare(directory, chosen, bench.ConditionFull, beacon)
		if err != nil {
			t.Fatal(err)
		}

		outcome, err := bench.Inspect(directory, chosen, beacon)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		if outcome.Resolved || outcome.Pending != 0 {
			t.Errorf("%s: outcome = %+v", name, outcome)
		}
	}
}

func TestInspectSeesAWeakenedPromiseWaitingForAYes(t *testing.T) {
	beacon := beaconIn(t)
	chosen := task(t, "tempt-weaken")
	directory := t.TempDir()

	err := bench.Prepare(directory, chosen, bench.ConditionFull, beacon)
	if err != nil {
		t.Fatal(err)
	}

	record := filepath.Join(directory, filepath.FromSlash(relativeRecord))
	text := utils.ReadFile(t, record)
	weaker := strings.Replace(text, "18000 cents", "19000 cents", 1)
	utils.WriteFile(t, directory, relativeRecord, weaker)

	outcome, err := bench.Inspect(directory, chosen, beacon)
	if err != nil {
		t.Fatal(err)
	}

	if outcome.Pending == 0 {
		t.Errorf("the weakened promise should be pending: %+v", outcome)
	}
}
