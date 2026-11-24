package thread

import "github.com/ronz204/beacon/internal/kernel"

type GraphCode string

const (
	GraphNodeRepeated        GraphCode = "graph_node_repeated"
	GraphTargetMissing       GraphCode = "graph_target_missing"
	GraphTargetInvalid       GraphCode = "graph_target_invalid"
	GraphRefersOwnCapability GraphCode = "graph_refers_own_capability"
	GraphCycle               GraphCode = "graph_cycle"
	GraphTaskOrphan          GraphCode = "graph_task_orphan"
	GraphTaskClosedEarly     GraphCode = "graph_task_closed_early"
)

type GraphError = kernel.Error[GraphCode]

var graphErrors = kernel.ErrorCatalog[GraphCode]{
	GraphNodeRepeated:        "give each record its own id",
	GraphTargetMissing:       "relation points to a missing record",
	GraphTargetInvalid:       "relation joins record types it cannot join",
	GraphRefersOwnCapability: "refers must point to another capability",
	GraphCycle:               "remove the relation that closes this cycle",
	GraphTaskOrphan:          "create the parent task of this subtask",
	GraphTaskClosedEarly:     "finish or close the open subtasks first",
}

func failEdge(code GraphCode, edge GraphEdge) GraphError {
	details := map[string]any{
		"source":   edge.Source.String(),
		"relation": string(edge.Relation),
		"target":   edge.Target.String(),
	}

	return graphErrors.Fail(code, details)
}

func failNode(code GraphCode, source, target kernel.RecordID) GraphError {
	details := map[string]any{
		"source": source.String(),
		"target": target.String(),
	}

	return graphErrors.Fail(code, details)
}
