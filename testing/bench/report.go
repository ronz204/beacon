package bench

import (
	"fmt"
	"math"
	"strings"
)

const confidenceZ = 1.96

type Result struct {
	Condition Condition
	Task      string
	Session   Session
}

type Rate struct {
	Hits  int
	Total int
}

func (r Rate) Value() float64 {
	if r.Total == 0 {
		return 0
	}

	return float64(r.Hits) / float64(r.Total)
}

func (r Rate) Interval() (float64, float64) {
	if r.Total == 0 {
		return 0, 1
	}

	n := float64(r.Total)
	p := r.Value()
	z2 := confidenceZ * confidenceZ

	center := (p + z2/(2*n)) / (1 + z2/n)
	spread := confidenceZ * math.Sqrt(p*(1-p)/n+z2/(4*n*n)) / (1 + z2/n)

	return math.Max(0, center-spread), math.Min(1, center+spread)
}

type Summary struct {
	Condition         Condition
	Sessions          int
	BriefAtStart      Rate
	ContextBeforeEdit Rate
	Resolved          Rate
	PendingAtEnd      Rate
	TamperPerSession  float64
	BlocksPerSession  float64
	MeanInjectedLines float64
	MeanCostUSD       float64
}

func Summarize(results []Result) []Summary {
	var summaries []Summary

	for _, condition := range Conditions() {
		var chosen []Session
		for _, result := range results {
			if result.Condition == condition {
				chosen = append(chosen, result.Session)
			}
		}
		if len(chosen) > 0 {
			summaries = append(summaries, summarize(condition, chosen))
		}
	}

	return summaries
}

func summarize(condition Condition, sessions []Session) Summary {
	summary := Summary{Condition: condition, Sessions: len(sessions)}
	count := float64(len(sessions))

	var tamper, blocks, lines int
	var cost float64

	for _, session := range sessions {
		summary.BriefAtStart.Total++
		if session.BriefAtStart() {
			summary.BriefAtStart.Hits++
		}

		summary.ContextBeforeEdit.Total += len(session.Edits)
		summary.ContextBeforeEdit.Hits += session.EditsWithContext()

		if session.Outcome != nil {
			summary.Resolved.Total++
			summary.PendingAtEnd.Total++

			if session.Outcome.Resolved {
				summary.Resolved.Hits++
			}
			if session.Outcome.Pending > 0 {
				summary.PendingAtEnd.Hits++
			}
		}

		tamper += session.Tampering
		blocks += session.BlockedByBeacon + session.BlockedByPermission
		lines += session.InjectedLines
		cost += session.CostUSD
	}

	summary.TamperPerSession = float64(tamper) / count
	summary.BlocksPerSession = float64(blocks) / count
	summary.MeanInjectedLines = float64(lines) / count
	summary.MeanCostUSD = cost / count

	return summary
}

func Format(summaries []Summary) string {
	var out strings.Builder

	out.WriteString("cond sessions brief@start ctx-before-edit ")
	out.WriteString("resolved pending@end tamper/s blocks/s lines/s cost/s\n")

	for _, summary := range summaries {
		fmt.Fprintf(
			&out,
			"%-4s %-8d %-11s %-15s %-8s %-11s %-8.2f %-8.2f %-7.1f %.4f\n",
			summary.Condition,
			summary.Sessions,
			percent(summary.BriefAtStart),
			percent(summary.ContextBeforeEdit),
			percent(summary.Resolved),
			percent(summary.PendingAtEnd),
			summary.TamperPerSession,
			summary.BlocksPerSession,
			summary.MeanInjectedLines,
			summary.MeanCostUSD,
		)
	}

	out.WriteString("rates are hits/total; the interval is Wilson 95 %, ")
	out.WriteString("so read small samples as hints, not as causes\n")

	return out.String()
}

func percent(rate Rate) string {
	if rate.Total == 0 {
		return "n/a"
	}

	low, high := rate.Interval()

	return fmt.Sprintf(
		"%d/%d %.0f-%.0f%%",
		rate.Hits,
		rate.Total,
		low*100,
		high*100,
	)
}
