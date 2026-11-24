package thread

import (
	"cmp"
	"maps"
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
)

const (
	visitNone = iota
	visitOpen
	visitDone
)

var cyclicRelations = []RelationKind{
	RelationBlocks,
	RelationParent,
	RelationSupersedes,
}

var allowedTargets = map[RelationKind]map[NodeKind][]NodeKind{
	RelationImplements: {
		NodeWork: {NodePromise},
		NodeTask: {NodePromise},
	},
	RelationBlocks: {
		NodeQuestion: {NodeWork, NodeTask},
		NodeTask:     {NodeTask},
	},
	RelationParent: {
		NodeTask: {NodeWork},
	},
	RelationSupersedes: {
		NodeDecision: {NodeDecision},
	},
	RelationServes: {
		NodeCapability: {NodeGoal},
		NodeWork:       {NodeGoal},
	},
	RelationOrigin: {
		NodeQuestion: {NodeWork},
		NodePromise:  {NodeWork},
	},
	RelationRefers: {
		NodePromise: {NodePromise},
		NodeGuide:   {NodeCapability},
	},
}

func compareIDs(a, b kernel.RecordID) int {
	return cmp.Compare(a.String(), b.String())
}

func compareEdges(a, b GraphEdge) int {
	return cmp.Or(
		compareIDs(a.Source, b.Source),
		cmp.Compare(a.Relation, b.Relation),
		compareIDs(a.Target, b.Target),
	)
}

func compareNodes(a, b GraphNode) int {
	return compareIDs(a.ID, b.ID)
}

func indexNodes(
	nodes []GraphNode,
) (map[kernel.RecordID]GraphNode, []error) {
	ordered := slices.Clone(nodes)
	slices.SortStableFunc(ordered, compareNodes)

	index := make(map[kernel.RecordID]GraphNode, len(ordered))
	var failures []error

	for _, node := range ordered {
		if _, repeated := index[node.ID]; repeated {
			failures = append(
				failures,
				failNode(GraphNodeRepeated, node.ID, node.ID),
			)
			continue
		}

		index[node.ID] = node
	}

	return index, failures
}

func validEdges(
	index map[kernel.RecordID]GraphNode,
	edges []GraphEdge,
) ([]GraphEdge, []error) {
	ordered := slices.Clone(edges)
	slices.SortFunc(ordered, compareEdges)
	ordered = slices.Compact(ordered)

	valid := make([]GraphEdge, 0, len(ordered))
	var failures []error

	for _, edge := range ordered {
		if err := edgeFailure(index, edge); err != nil {
			failures = append(failures, err)
			continue
		}

		valid = append(valid, edge)
	}

	return valid, failures
}

func edgeFailure(index map[kernel.RecordID]GraphNode, edge GraphEdge) error {
	target, found := index[edge.Target]
	if !found {
		return failEdge(GraphTargetMissing, edge)
	}

	source := index[edge.Source]
	allowed := allowedTargets[edge.Relation][source.Kind]
	if !slices.Contains(allowed, target.Kind) {
		return failEdge(GraphTargetInvalid, edge)
	}

	if refersOwnCapability(edge, source, target) {
		return failEdge(GraphRefersOwnCapability, edge)
	}

	return nil
}

func refersOwnCapability(edge GraphEdge, source, target GraphNode) bool {
	if edge.Relation != RelationRefers || source.Kind != NodePromise {
		return false
	}

	return source.Owner == target.Owner
}

func cycleFailures(edges []GraphEdge) []error {
	var failures []error

	for _, relation := range cyclicRelations {
		related := slices.DeleteFunc(
			slices.Clone(edges),
			func(edge GraphEdge) bool {
				return edge.Relation != relation
			},
		)
		failures = append(failures, findCycles(related)...)
	}

	return failures
}

func findCycles(edges []GraphEdge) []error {
	next := make(map[kernel.RecordID][]GraphEdge)
	for _, edge := range edges {
		next[edge.Source] = append(next[edge.Source], edge)
	}

	visits := make(map[kernel.RecordID]int)
	var failures []error

	var visit func(kernel.RecordID)
	visit = func(node kernel.RecordID) {
		visits[node] = visitOpen

		for _, edge := range next[node] {
			switch visits[edge.Target] {
			case visitOpen:
				failures = append(failures, failEdge(GraphCycle, edge))
			case visitNone:
				visit(edge.Target)
			}
		}

		visits[node] = visitDone
	}

	for _, edge := range edges {
		if visits[edge.Source] == visitNone {
			visit(edge.Source)
		}
	}

	return failures
}

func hierarchyFailures(index map[kernel.RecordID]GraphNode) []error {
	nodes := slices.SortedFunc(maps.Values(index), compareNodes)

	var failures []error
	for _, node := range nodes {
		if node.Kind == NodeTask {
			failures = append(failures, taskFailures(index, node)...)
		}
	}

	return failures
}

func taskFailures(
	index map[kernel.RecordID]GraphNode,
	task GraphNode,
) []error {
	root := task.ID.Root()
	parent, found := task.ID.Parent()

	var failures []error
	if found && parent != root {
		if index[parent].Kind != NodeTask {
			failures = append(
				failures,
				failNode(GraphTaskOrphan, task.ID, parent),
			)
		}
	}

	if !task.Unfinished {
		return failures
	}

	for found && parent != root {
		ancestor, exists := index[parent]
		if exists && !ancestor.Unfinished {
			failures = append(
				failures,
				failNode(GraphTaskClosedEarly, parent, task.ID),
			)
		}

		parent, found = parent.Parent()
	}

	return failures
}
