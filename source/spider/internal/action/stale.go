package action

import (
	"context"
	"fmt"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

const (
	churnThreshold = 10
	decisionReview = "decision"
)

type StaleRecord struct {
	ID      kernel.RecordID
	Kind    thread.NodeKind
	Due     kernel.CalendarDate
	Overdue bool
	Commits int
}

type reviewSubject struct {
	id    fmt.Stringer
	kind  thread.NodeKind
	base  kernel.CalendarDate
	span  memory.ReviewSpan
	globs []kernel.PathGlob
}

type staleCheck struct {
	sources  Sources
	commits  []memory.CodeCommit
	failures []error
}

func staleRecords(
	ctx context.Context,
	snapshot Snapshot,
	sources Sources,
) ([]StaleRecord, []error) {
	subjects, failures := reviewSubjects(ctx, snapshot, sources)
	check := staleCheck{sources: sources, failures: failures}
	check.loadCommits(ctx, subjects)

	var stale []StaleRecord
	for _, subject := range subjects {
		record, isStale := check.evaluate(subject)
		if isStale {
			stale = append(stale, record)
		}
	}

	return stale, check.failures
}

func reviewSubjects(
	ctx context.Context,
	snapshot Snapshot,
	sources Sources,
) ([]reviewSubject, []error) {
	var subjects []reviewSubject

	for _, decision := range snapshot.Decisions {
		if decision.Status() == memory.DecisionActive {
			subjects = append(subjects, decisionSubject(decision, sources))
		}
	}

	guideDays, err := guideChangeDays(ctx, snapshot.Guides, sources)
	for _, guide := range snapshot.Guides {
		subject := guideSubject(guide, guideDays, sources)
		if !kernel.IsMissing(subject.base) {
			subjects = append(subjects, subject)
		}
	}

	if err != nil {
		return subjects, []error{err}
	}

	return subjects, nil
}

func decisionSubject(
	decision memory.Decision,
	sources Sources,
) reviewSubject {
	base, reviewed := decision.Reviewed()
	if !reviewed {
		base = decision.Created()
	}

	span, found := decision.Review()
	if !found {
		span = sources.Review[decisionReview]
	}

	return reviewSubject{
		id:    decision.ID(),
		kind:  thread.NodeDecision,
		base:  base,
		span:  span,
		globs: decision.Where(),
	}
}

func guideSubject(
	guide memory.Guide,
	days map[memory.GuideID]kernel.CalendarDate,
	sources Sources,
) reviewSubject {
	base, reviewed := guide.Reviewed()
	if !reviewed {
		base = days[guide.ID()]
	}

	span, found := guide.Review()
	if !found {
		span = sources.Review[string(guide.Kind())]
	}

	return reviewSubject{
		id:    guide.ID(),
		kind:  thread.NodeGuide,
		base:  base,
		span:  span,
		globs: guide.Applies(),
	}
}

func guideChangeDays(
	ctx context.Context,
	guides []memory.Guide,
	sources Sources,
) (map[memory.GuideID]kernel.CalendarDate, error) {
	unreviewed := slices.ContainsFunc(guides, func(guide memory.Guide) bool {
		_, reviewed := guide.Reviewed()
		return !reviewed
	})
	if !unreviewed {
		return nil, nil
	}

	return sources.History.GuideChangeDays(ctx)
}

func (c *staleCheck) loadCommits(
	ctx context.Context,
	subjects []reviewSubject,
) {
	var oldest kernel.CalendarDate
	for _, subject := range subjects {
		if len(subject.globs) == 0 {
			continue
		}

		if kernel.IsMissing(oldest) || subject.base.Before(oldest) {
			oldest = subject.base
		}
	}

	if kernel.IsMissing(oldest) {
		return
	}

	commits, err := c.sources.History.CommitsSince(ctx, oldest)
	if err != nil {
		c.failures = append(c.failures, err)
		return
	}

	c.commits = commits
}

func (c *staleCheck) evaluate(subject reviewSubject) (StaleRecord, bool) {
	id, err := recordIDOf(subject.id)
	if err != nil {
		return StaleRecord{}, false
	}

	record := StaleRecord{
		ID:      id,
		Kind:    subject.kind,
		Commits: c.churn(subject),
	}

	if !kernel.IsMissing(subject.span) {
		record.Due = subject.span.DueAfter(subject.base)
		record.Overdue = record.Due.Before(c.sources.Today)
	}

	return record, record.Overdue || record.Commits >= churnThreshold
}

func (c *staleCheck) churn(subject reviewSubject) int {
	count := 0

	for _, commit := range c.commits {
		if commit.Day.Before(subject.base) {
			continue
		}

		touched, err := globsTouch(
			c.sources.Matcher,
			subject.globs,
			commit.Files,
		)
		if err != nil {
			c.failures = append(c.failures, err)
			continue
		}

		if touched {
			count++
		}
	}

	return count
}

func staleSignals(records []StaleRecord) []thread.Signal {
	signals := make([]thread.Signal, 0, len(records))

	for _, record := range records {
		signals = append(signals, thread.Signal{
			Kind:    thread.SignalStale,
			Subject: record.ID,
		})
	}

	return signals
}
