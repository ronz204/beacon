package action

import (
	"context"
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

const findLimit = 20

type FindMatch struct {
	Kind thread.NodeKind
	ID   string
	Text string
}

type FindResult struct {
	Matches []FindMatch
	Omitted int
}

type findSearch struct {
	needle  string
	matches []FindMatch
}

func Find(
	ctx context.Context,
	repositories Repositories,
	text string,
) (FindResult, []error) {
	snapshot := Load(ctx, repositories)
	needle := kernel.FoldText(strings.TrimSpace(text))
	if needle == "" {
		return FindResult{}, snapshot.Failures
	}

	search := &findSearch{needle: needle}
	search.snapshot(snapshot)

	return search.result(), snapshot.Failures
}

func (s *findSearch) snapshot(snapshot Snapshot) {
	for _, goal := range snapshot.Goals {
		s.add(thread.NodeGoal, goal.ID(), goal.Statement(), goal.Success())
	}
	for _, capability := range snapshot.Capabilities {
		s.add(
			thread.NodeCapability,
			capability.ID(),
			capability.Name(),
			capability.Intent(),
		)
		for _, promise := range capability.Promises() {
			s.add(thread.NodePromise, promise.ID(), promise.Text())
		}
	}
	for _, decision := range snapshot.Decisions {
		s.add(
			thread.NodeDecision,
			decision.ID(),
			decision.What(),
			decision.Why(),
		)
	}
	for _, question := range snapshot.Questions {
		s.add(thread.NodeQuestion, question.ID(), question.Text())
	}
	for _, guide := range snapshot.Guides {
		s.add(thread.NodeGuide, guide.ID(), guide.Title())
	}
	for _, work := range snapshot.Works {
		s.add(thread.NodeWork, work.ID(), work.Request())
	}
	for _, task := range snapshot.Tasks {
		s.add(thread.NodeTask, task.ID(), task.Title())
	}
}

func (s *findSearch) add(
	kind thread.NodeKind,
	id fmt.Stringer,
	texts ...fmt.Stringer,
) {
	for _, text := range texts {
		folded := kernel.FoldText(text.String())
		if strings.Contains(folded, s.needle) {
			s.matches = append(s.matches, FindMatch{
				Kind: kind,
				ID:   id.String(),
				Text: firstLine(texts[0].String()),
			})
			return
		}
	}
}

func (s *findSearch) result() FindResult {
	if len(s.matches) <= findLimit {
		return FindResult{Matches: s.matches}
	}

	return FindResult{
		Matches: s.matches[:findLimit],
		Omitted: len(s.matches) - findLimit,
	}
}
