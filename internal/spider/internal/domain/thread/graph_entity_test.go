package thread_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/thread"
)

const (
	goalID       = "bc-0001"
	capabilityID = "bc-7f3a"
	otherCapID   = "bc-7f3b"
	promiseID    = "bc-c3d4"
	otherPromID  = "bc-c3d5"
	foreignPrmID = "bc-c3d6"
	decisionID   = "bc-a1b2"
	otherDecID   = "bc-a1b3"
	questionID   = "bc-e5f6"
	guideID      = "bc-0a9b"
	workID       = "bc-91cd"
	taskID       = "bc-91cd.1"
	otherTaskID  = "bc-91cd.2"
	subtaskID    = "bc-91cd.1.1"
)

const (
	relImplements = thread.RelationImplements
	relBlocks     = thread.RelationBlocks
	relParent     = thread.RelationParent
	relSupersedes = thread.RelationSupersedes
	relServes     = thread.RelationServes
	relOrigin     = thread.RelationOrigin
	relRefers     = thread.RelationRefers
)

func recordID(raw string) kernel.RecordID {
	id, err := kernel.ParseRecordID(raw)
	if err != nil {
		panic(err)
	}

	return id
}

func node(raw string, kind thread.NodeKind) thread.GraphNode {
	return thread.GraphNode{ID: recordID(raw), Kind: kind}
}

func promise(raw, owner string) thread.GraphNode {
	return thread.GraphNode{
		ID:    recordID(raw),
		Kind:  thread.NodePromise,
		Owner: recordID(owner),
	}
}

func task(raw string, unfinished bool) thread.GraphNode {
	return thread.GraphNode{
		ID:         recordID(raw),
		Kind:       thread.NodeTask,
		Unfinished: unfinished,
	}
}

func edge(
	source string,
	relation thread.RelationKind,
	target string,
) thread.GraphEdge {
	return thread.GraphEdge{
		Source:   recordID(source),
		Relation: relation,
		Target:   recordID(target),
	}
}

func sampleNodes() []thread.GraphNode {
	return []thread.GraphNode{
		node(goalID, thread.NodeGoal),
		node(capabilityID, thread.NodeCapability),
		node(otherCapID, thread.NodeCapability),
		promise(promiseID, capabilityID),
		promise(otherPromID, capabilityID),
		promise(foreignPrmID, otherCapID),
		node(decisionID, thread.NodeDecision),
		node(otherDecID, thread.NodeDecision),
		node(questionID, thread.NodeQuestion),
		node(guideID, thread.NodeGuide),
		node(workID, thread.NodeWork),
		task(taskID, true),
		task(otherTaskID, true),
	}
}

func requireCodes(
	t *testing.T,
	failures []error,
	want ...thread.GraphCode,
) {
	t.Helper()

	got := make([]thread.GraphCode, 0, len(failures))
	for _, failure := range failures {
		var located thread.GraphError
		if !errors.As(failure, &located) {
			t.Fatalf("expected GraphError, got %v", failure)
		}
		got = append(got, located.Code)
	}

	if !slices.Equal(got, want) {
		t.Fatalf("codes = %v, want %v", got, want)
	}
}

func TestNewGraphAcceptsEveryValidRelation(t *testing.T) {
	edges := []thread.GraphEdge{
		edge(workID, relImplements, promiseID),
		edge(taskID, relImplements, promiseID),
		edge(questionID, relBlocks, workID),
		edge(questionID, relBlocks, taskID),
		edge(taskID, relBlocks, otherTaskID),
		edge(taskID, relParent, workID),
		edge(decisionID, relSupersedes, otherDecID),
		edge(capabilityID, relServes, goalID),
		edge(workID, relServes, goalID),
		edge(questionID, relOrigin, workID),
		edge(promiseID, relOrigin, workID),
		edge(promiseID, relRefers, foreignPrmID),
		edge(guideID, relRefers, capabilityID),
	}

	_, failures := thread.NewGraph(sampleNodes(), edges)

	requireCodes(t, failures)
}

func TestNewGraphRejectsInvalidRelations(t *testing.T) {
	cases := []struct {
		name string
		edge thread.GraphEdge
	}{
		{"implements goal", edge(workID, relImplements, goalID)},
		{"implements from guide", edge(guideID, relImplements, promiseID)},
		{"blocks goal", edge(questionID, relBlocks, goalID)},
		{"task blocks work", edge(taskID, relBlocks, workID)},
		{"parent task", edge(taskID, relParent, otherTaskID)},
		{"supersedes goal", edge(decisionID, relSupersedes, goalID)},
		{"serves work", edge(capabilityID, relServes, workID)},
		{"origin task", edge(questionID, relOrigin, taskID)},
		{"promise refers guide", edge(promiseID, relRefers, guideID)},
		{"guide refers goal", edge(guideID, relRefers, goalID)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			edges := []thread.GraphEdge{c.edge}

			_, failures := thread.NewGraph(sampleNodes(), edges)

			requireCodes(t, failures, thread.GraphTargetInvalid)
		})
	}
}

func TestNewGraphRejectsRepeatedID(t *testing.T) {
	nodes := append(sampleNodes(), node(goalID, thread.NodeDecision))

	_, failures := thread.NewGraph(nodes, nil)

	requireCodes(t, failures, thread.GraphNodeRepeated)
}

func TestNewGraphRejectsMissingTarget(t *testing.T) {
	edges := []thread.GraphEdge{
		edge(capabilityID, relServes, "bc-ffff"),
	}

	_, failures := thread.NewGraph(sampleNodes(), edges)

	requireCodes(t, failures, thread.GraphTargetMissing)
}

func TestNewGraphRejectsRefersInsideCapability(t *testing.T) {
	edges := []thread.GraphEdge{
		edge(promiseID, relRefers, otherPromID),
	}

	_, failures := thread.NewGraph(sampleNodes(), edges)

	requireCodes(t, failures, thread.GraphRefersOwnCapability)
}

func TestNewGraphRejectsCycles(t *testing.T) {
	cases := []struct {
		name  string
		edges []thread.GraphEdge
	}{
		{
			"blocks",
			[]thread.GraphEdge{
				edge(taskID, relBlocks, otherTaskID),
				edge(otherTaskID, relBlocks, taskID),
			},
		},
		{
			"blocks itself",
			[]thread.GraphEdge{
				edge(taskID, relBlocks, taskID),
			},
		},
		{
			"supersedes",
			[]thread.GraphEdge{
				edge(decisionID, relSupersedes, otherDecID),
				edge(otherDecID, relSupersedes, decisionID),
			},
		},
		{
			"supersedes itself",
			[]thread.GraphEdge{
				edge(decisionID, relSupersedes, decisionID),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, failures := thread.NewGraph(sampleNodes(), c.edges)

			requireCodes(t, failures, thread.GraphCycle)
		})
	}
}

func TestNewGraphAcceptsSharedDependencyWithoutCycle(t *testing.T) {
	nodes := append(
		sampleNodes(),
		task("bc-91cd.3", true),
		task("bc-91cd.4", true),
	)
	edges := []thread.GraphEdge{
		edge(taskID, relBlocks, otherTaskID),
		edge(taskID, relBlocks, "bc-91cd.3"),
		edge(otherTaskID, relBlocks, "bc-91cd.4"),
		edge("bc-91cd.3", relBlocks, "bc-91cd.4"),
	}

	_, failures := thread.NewGraph(nodes, edges)

	requireCodes(t, failures)
}

func TestNewGraphRejectsOrphanSubtask(t *testing.T) {
	nodes := append(sampleNodes(), task("bc-91cd.9.1", true))

	_, failures := thread.NewGraph(nodes, nil)

	requireCodes(t, failures, thread.GraphTaskOrphan)
}

func TestNewGraphAcceptsSubtaskWithParent(t *testing.T) {
	nodes := append(sampleNodes(), task(subtaskID, true))

	_, failures := thread.NewGraph(nodes, nil)

	requireCodes(t, failures)
}

func TestNewGraphRejectsDoneTaskWithOpenSubtask(t *testing.T) {
	nodes := []thread.GraphNode{
		node(workID, thread.NodeWork),
		task(taskID, false),
		task(subtaskID, true),
	}

	_, failures := thread.NewGraph(nodes, nil)

	requireCodes(t, failures, thread.GraphTaskClosedEarly)
}

func TestNewGraphChecksEveryAncestor(t *testing.T) {
	nodes := []thread.GraphNode{
		node(workID, thread.NodeWork),
		task(taskID, false),
		task(subtaskID, false),
		task("bc-91cd.1.1.1", true),
	}

	_, failures := thread.NewGraph(nodes, nil)

	requireCodes(
		t,
		failures,
		thread.GraphTaskClosedEarly,
		thread.GraphTaskClosedEarly,
	)
}

func TestNewGraphAcceptsDoneTaskWithDoneSubtask(t *testing.T) {
	nodes := []thread.GraphNode{
		node(workID, thread.NodeWork),
		task(taskID, false),
		task(subtaskID, false),
	}

	_, failures := thread.NewGraph(nodes, nil)

	requireCodes(t, failures)
}

func TestNewGraphReportsEveryFailureInStableOrder(t *testing.T) {
	edges := []thread.GraphEdge{
		edge(workID, relServes, "bc-ffff"),
		edge(capabilityID, relServes, "bc-eeee"),
	}
	reversed := slices.Clone(edges)
	slices.Reverse(reversed)

	_, first := thread.NewGraph(sampleNodes(), edges)
	_, second := thread.NewGraph(sampleNodes(), reversed)

	requireCodes(
		t,
		first,
		thread.GraphTargetMissing,
		thread.GraphTargetMissing,
	)

	for index := range first {
		var left, right thread.GraphError
		errors.As(first[index], &left)
		errors.As(second[index], &right)

		if left.Details["source"] != right.Details["source"] {
			t.Fatalf("order changed: %v, %v", left.Details, right.Details)
		}
	}
}

func TestNeighborsAreDeterministic(t *testing.T) {
	edges := []thread.GraphEdge{
		edge(workID, relImplements, promiseID),
		edge(taskID, relImplements, promiseID),
		edge(questionID, relBlocks, taskID),
		edge(capabilityID, relServes, goalID),
		edge(taskID, relImplements, promiseID),
	}
	reversed := slices.Clone(edges)
	slices.Reverse(reversed)

	first, failures := thread.NewGraph(sampleNodes(), edges)
	requireCodes(t, failures)
	second, failures := thread.NewGraph(sampleNodes(), reversed)
	requireCodes(t, failures)

	got := first.Neighbors(recordID(promiseID))
	want := second.Neighbors(recordID(promiseID))

	if len(got) != 2 {
		t.Fatalf("neighbors = %v", got)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("order changed: %v, %v", got, want)
	}
	if neighbors := first.Neighbors(recordID(guideID)); len(neighbors) != 0 {
		t.Fatalf("guide neighbors = %v", neighbors)
	}
}
