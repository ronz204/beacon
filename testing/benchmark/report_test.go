package benchmark_test

import (
	"strings"
	"testing"

	"github.com/ronz204/beacon/testing/benchmark"
)

func TestSummaryTellsTheTwoTranscriptsApartInEveryMetric(t *testing.T) {
	results := []benchmark.Result{
		{Condition: benchmark.ConditionFull, Session: replay(t, "ideal")},
		{Condition: benchmark.ConditionWithout, Session: replay(t, "ignoring")},
	}

	summaries := benchmark.Summarize(results)
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
	small := benchmark.Rate{Hits: 1, Total: 2}
	large := benchmark.Rate{Hits: 50, Total: 100}

	smallLow, smallHigh := small.Interval()
	largeLow, largeHigh := large.Interval()

	if smallHigh-smallLow <= largeHigh-largeLow {
		t.Errorf("small %v-%v, large %v-%v",
			smallLow, smallHigh, largeLow, largeHigh)
	}

	empty := benchmark.Rate{}
	if low, high := empty.Interval(); low != 0 || high != 1 {
		t.Errorf("empty interval = %v-%v", low, high)
	}
}

func TestFormatNamesEveryColumnAndTheCaveat(t *testing.T) {
	results := []benchmark.Result{
		{Condition: benchmark.ConditionFull, Session: replay(t, "ideal")},
	}

	text := benchmark.Format(benchmark.Summarize(results))

	for _, want := range []string{"brief@start", "ctx-before-edit", "Wilson"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
}
