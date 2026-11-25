package thread

import (
	"slices"

	"github.com/ronz204/beacon/source/kernel"
)

type Graph struct {
	nodes map[kernel.RecordID]GraphNode
	edges []GraphEdge
}

func NewGraph(nodes []GraphNode, edges []GraphEdge) (Graph, []error) {
	index, failures := indexNodes(nodes)

	valid, invalid := validEdges(index, edges)
	failures = append(failures, invalid...)
	failures = append(failures, cycleFailures(valid)...)
	failures = append(failures, hierarchyFailures(index)...)

	return Graph{nodes: index, edges: valid}, failures
}

func (g Graph) Node(id kernel.RecordID) (GraphNode, bool) {
	node, found := g.nodes[id]
	return node, found
}

func (g Graph) Neighbors(id kernel.RecordID) []GraphEdge {
	return slices.DeleteFunc(
		slices.Clone(g.edges),
		func(edge GraphEdge) bool {
			return edge.Source != id && edge.Target != id
		},
	)
}
