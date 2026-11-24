package thread

import (
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
)

type Graph struct {
	edges []GraphEdge
}

func NewGraph(nodes []GraphNode, edges []GraphEdge) (Graph, []error) {
	index, failures := indexNodes(nodes)

	valid, invalid := validEdges(index, edges)
	failures = append(failures, invalid...)
	failures = append(failures, cycleFailures(valid)...)
	failures = append(failures, hierarchyFailures(index)...)

	return Graph{edges: valid}, failures
}

func (g Graph) Neighbors(id kernel.RecordID) []GraphEdge {
	return slices.DeleteFunc(
		slices.Clone(g.edges),
		func(edge GraphEdge) bool {
			return edge.Source != id && edge.Target != id
		},
	)
}
