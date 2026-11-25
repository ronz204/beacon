package action

import (
	"context"
	"fmt"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

type PromiseContext struct {
	Promise      ledger.Promise
	Verification ledger.Verification
}

type CapabilityContext struct {
	Capability ledger.Capability
	Promises   []PromiseContext
}

type ContextResult struct {
	Goals        []ledger.Goal
	Capabilities []CapabilityContext
	Decisions    []memory.Decision
	Guides       []memory.Guide
	Questions    []memory.Question
	Signals      []thread.Signal
}

type contextSelector struct {
	matcher  kernel.PathMatcher
	paths    []string
	graph    thread.Graph
	touched  map[kernel.RecordID]bool
	linked   map[kernel.RecordID]bool
	failures []error
}

func Context(
	ctx context.Context,
	sources Sources,
	paths []string,
) (ContextResult, []error) {
	analysis, failures := Analyze(ctx, sources)
	snapshot := analysis.Snapshot

	selector := &contextSelector{
		matcher: sources.Matcher,
		paths:   paths,
		graph:   analysis.Graph,
		touched: make(map[kernel.RecordID]bool),
		linked:  make(map[kernel.RecordID]bool),
	}

	result := ContextResult{Goals: activeGoals(snapshot.Goals)}
	result.Capabilities = selector.capabilities(
		snapshot.Capabilities,
		analysis.Verifications,
	)
	result.Decisions = selector.decisions(snapshot.Decisions)
	result.Guides = selector.guides(snapshot.Guides)
	selector.linkWork()
	result.Questions = selector.questions(snapshot.Questions)
	result.Signals = selector.signals(analysis.Signals)

	return result, slices.Concat(failures, selector.failures)
}

func activeGoals(goals []ledger.Goal) []ledger.Goal {
	return slices.DeleteFunc(slices.Clone(goals), func(goal ledger.Goal) bool {
		return goal.Status() != ledger.GoalActive
	})
}

func (s *contextSelector) matches(globs []kernel.PathGlob) bool {
	for _, glob := range globs {
		for _, path := range s.paths {
			matched, err := s.matcher.MatchPath(glob, path)
			if err != nil {
				s.failures = append(s.failures, err)
				continue
			}

			if matched {
				return true
			}
		}
	}

	return false
}

func (s *contextSelector) touch(id fmt.Stringer) {
	if parsed, err := recordIDOf(id); err == nil {
		s.touched[parsed] = true
	}
}

func (s *contextSelector) isTouched(id fmt.Stringer) bool {
	parsed, err := recordIDOf(id)
	return err == nil && s.touched[parsed]
}

func (s *contextSelector) capabilities(
	capabilities []ledger.Capability,
	verifications Verifications,
) []CapabilityContext {
	var selected []CapabilityContext

	for _, capability := range capabilities {
		if !s.matches(capability.Paths()) {
			continue
		}

		s.touch(capability.ID())
		promises := make([]PromiseContext, 0, len(capability.Promises()))
		for _, promise := range capability.Promises() {
			s.touch(promise.ID())
			promises = append(promises, PromiseContext{
				Promise:      promise,
				Verification: verifications[promise.ID()],
			})
		}

		selected = append(selected, CapabilityContext{
			Capability: capability,
			Promises:   promises,
		})
	}

	return selected
}

func (s *contextSelector) decisions(
	decisions []memory.Decision,
) []memory.Decision {
	var selected []memory.Decision

	for _, decision := range decisions {
		active := decision.Status() == memory.DecisionActive
		if active && s.matches(decision.Where()) {
			s.touch(decision.ID())
			selected = append(selected, decision)
		}
	}

	return selected
}

func (s *contextSelector) guides(guides []memory.Guide) []memory.Guide {
	var selected []memory.Guide

	for _, guide := range guides {
		related := s.relatesGuide(guide)
		if related {
			s.touch(guide.ID())
		}

		if related || guide.Always() {
			selected = append(selected, guide)
		}
	}

	return selected
}

func (s *contextSelector) relatesGuide(guide memory.Guide) bool {
	if s.matches(guide.Applies()) {
		return true
	}

	return slices.ContainsFunc(
		guide.Refers(),
		func(capability memory.GuideCapabilityID) bool {
			return s.isTouched(capability)
		},
	)
}

func (s *contextSelector) linkWork() {
	for id := range s.touched {
		for _, edge := range s.graph.Neighbors(id) {
			other := otherEnd(edge, id)
			node, found := s.graph.Node(other)

			if found && isWorkNode(node) {
				s.linked[other] = true
			}
		}
	}
}

func isWorkNode(node thread.GraphNode) bool {
	return node.Kind == thread.NodeWork || node.Kind == thread.NodeTask
}

func otherEnd(edge thread.GraphEdge, id kernel.RecordID) kernel.RecordID {
	if edge.Source == id {
		return edge.Target
	}

	return edge.Source
}

func (s *contextSelector) inScope(id kernel.RecordID) bool {
	return s.touched[id] || s.linked[id]
}

func (s *contextSelector) questions(
	questions []memory.Question,
) []memory.Question {
	var selected []memory.Question

	for _, question := range questions {
		if question.Status() != memory.QuestionOpen {
			continue
		}

		id, err := recordIDOf(question.ID())
		if err != nil {
			continue
		}

		related := slices.ContainsFunc(
			s.graph.Neighbors(id),
			func(edge thread.GraphEdge) bool {
				return s.inScope(otherEnd(edge, id))
			},
		)
		if related {
			selected = append(selected, question)
		}
	}

	return selected
}

func (s *contextSelector) signals(all []thread.Signal) []thread.Signal {
	selected := slices.DeleteFunc(slices.Clone(all), func(
		signal thread.Signal,
	) bool {
		return !s.inScope(signal.Subject)
	})

	touched := make([]kernel.RecordID, 0, len(s.touched))
	for id := range s.touched {
		touched = append(touched, id)
	}

	selected = append(selected, thread.TensionSignals(s.graph, touched)...)
	slices.SortFunc(selected, thread.CompareSignals)

	return slices.Compact(selected)
}
