package action

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

type BriefSection string

const (
	BriefGoal     BriefSection = "goal"
	BriefBlockers BriefSection = "blockers"
	BriefProgress BriefSection = "progress"
	BriefQuestion BriefSection = "question"
	BriefPending  BriefSection = "pending"
	BriefChecks   BriefSection = "checks"
	BriefGuides   BriefSection = "guides"
	BriefOverflow BriefSection = "overflow"
)

type BriefLine struct {
	Section BriefSection
	Text    string
}

type BriefResult struct {
	Lines   []BriefLine
	Omitted int
	Signals []thread.Signal
}

type briefTexts struct {
	questions map[string]string
	promises  map[string]string
	reviews   map[string]string
}

func Brief(
	ctx context.Context,
	sources Sources,
	limit int,
	pending []string,
) (BriefResult, []error) {
	analysis, failures := Analyze(ctx, sources)
	snapshot := analysis.Snapshot
	texts := newBriefTexts(snapshot)

	ranked := slices.Concat(
		texts.blockerLines(analysis.Signals),
		progressLines(snapshot),
		pendingLines(pending),
		texts.checkLines(analysis.Signals, analysis.Stale),
		openQuestionLines(snapshot.Questions, analysis.Signals),
		guideLines(snapshot.Guides),
	)

	result := fitBrief(goalLines(snapshot.Goals), ranked, limit)
	result.Signals = analysis.Signals

	return result, failures
}

func newBriefTexts(snapshot Snapshot) briefTexts {
	texts := briefTexts{
		questions: make(map[string]string),
		promises:  make(map[string]string),
		reviews:   make(map[string]string),
	}

	for _, question := range snapshot.Questions {
		texts.questions[question.ID().String()] = question.Text().String()
	}

	for _, decision := range snapshot.Decisions {
		texts.reviews[decision.ID().String()] = decision.What().String()
	}

	for _, guide := range snapshot.Guides {
		texts.reviews[guide.ID().String()] = guide.Title().String()
	}

	for _, capability := range snapshot.Capabilities {
		for _, promise := range capability.Promises() {
			id := promise.ID().String()
			texts.promises[id] = promise.Text().String()
		}
	}

	return texts
}

func fitBrief(fixed, ranked []BriefLine, limit int) BriefResult {
	if len(fixed)+len(ranked) <= limit {
		return BriefResult{Lines: slices.Concat(fixed, ranked)}
	}

	room := min(max(limit-len(fixed)-1, 0), len(ranked))
	omitted := len(ranked) - room
	overflow := BriefLine{
		Section: BriefOverflow,
		Text:    fmt.Sprintf("%d more, run beacon status", omitted),
	}

	return BriefResult{
		Lines:   slices.Concat(fixed, ranked[:room], []BriefLine{overflow}),
		Omitted: omitted,
	}
}

func goalLines(goals []ledger.Goal) []BriefLine {
	var lines []BriefLine

	for _, goal := range activeGoals(goals) {
		lines = append(lines, briefLine(
			BriefGoal,
			"Goal %s: %s",
			goal.ID(),
			firstLine(goal.Statement().String()),
		))

		for _, nongoal := range goal.Nongoals() {
			lines = append(lines, briefLine(
				BriefGoal,
				"Not doing: %s",
				firstLine(nongoal.String()),
			))
		}
	}

	return lines
}

func (t briefTexts) blockerLines(signals []thread.Signal) []BriefLine {
	var lines []BriefLine

	for _, signal := range signals {
		switch signal.Kind {
		case thread.SignalBlocked:
			lines = append(lines, briefLine(
				BriefBlockers,
				"Blocked: %s by question %s: %s",
				signal.Subject,
				signal.Related,
				firstLine(t.questions[signal.Related.String()]),
			))
		case thread.SignalRed:
			lines = append(lines, briefLine(
				BriefBlockers,
				"Red: promise %s: %s",
				signal.Subject,
				firstLine(t.promises[signal.Subject.String()]),
			))
		}
	}

	return lines
}

func progressLines(snapshot Snapshot) []BriefLine {
	var lines []BriefLine

	for _, work := range snapshot.Works {
		if work.Status() == agenda.WorkActive {
			lines = append(lines, briefLine(
				BriefProgress,
				"Work %s: %s",
				work.ID(),
				firstLine(work.Request().String()),
			))
		}
	}

	for _, task := range snapshot.Tasks {
		if task.Status() == agenda.TaskClaimed {
			lines = append(lines, claimedTaskLine(task))
		}
	}

	return lines
}

func claimedTaskLine(task agenda.Task) BriefLine {
	branch, _ := task.Branch()

	return briefLine(
		BriefProgress,
		"Task %s claimed on %s: %s",
		task.ID(),
		branch,
		firstLine(task.Title().String()),
	)
}

func pendingLines(pending []string) []BriefLine {
	lines := make([]BriefLine, 0, len(pending))

	for _, text := range pending {
		lines = append(lines, briefLine(
			BriefPending,
			"Needs approval: %s",
			firstLine(text),
		))
	}

	return lines
}

func (t briefTexts) checkLines(
	signals []thread.Signal,
	stale []StaleRecord,
) []BriefLine {
	var lines []BriefLine
	pending := 0

	for _, signal := range signals {
		switch signal.Kind {
		case thread.SignalDrift:
			lines = append(lines, briefLine(
				BriefChecks,
				"Drift: promise %s: %s",
				signal.Subject,
				firstLine(t.promises[signal.Subject.String()]),
			))
		case thread.SignalNotRunHere:
			pending++
		}
	}

	if pending > 0 {
		lines = append(lines, briefLine(
			BriefChecks,
			"Promises without a check result on this machine: %d",
			pending,
		))
	}

	for _, record := range stale {
		lines = append(lines, briefLine(
			BriefChecks,
			"Stale: %s %s: %s",
			record.Kind,
			record.ID,
			firstLine(t.reviews[record.ID.String()]),
		))
	}

	return lines
}

func openQuestionLines(
	questions []memory.Question,
	signals []thread.Signal,
) []BriefLine {
	blocking := map[string]bool{}
	for _, signal := range signals {
		if signal.Kind == thread.SignalBlocked {
			blocking[signal.Related.String()] = true
		}
	}

	var lines []BriefLine
	for _, question := range questions {
		id := question.ID().String()
		if question.Status() == memory.QuestionOpen && !blocking[id] {
			lines = append(lines, briefLine(
				BriefQuestion,
				"Open question %s: %s",
				id,
				firstLine(question.Text().String()),
			))
		}
	}

	return lines
}

func guideLines(guides []memory.Guide) []BriefLine {
	lines := make([]BriefLine, 0, len(guides))

	for _, guide := range guides {
		lines = append(lines, briefLine(
			BriefGuides,
			"Guide %s (%s): %s",
			guide.ID(),
			guide.Kind(),
			firstLine(guide.Title().String()),
		))
	}

	return lines
}

func briefLine(section BriefSection, format string, args ...any) BriefLine {
	return BriefLine{Section: section, Text: fmt.Sprintf(format, args...)}
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}
