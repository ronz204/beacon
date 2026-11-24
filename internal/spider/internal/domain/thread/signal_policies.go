package thread

import (
	"cmp"

	"github.com/ronz204/beacon/internal/kernel"
)

func BlockedSignals(graph Graph) []Signal {
	var signals []Signal

	for _, edge := range graph.edges {
		if edge.Relation != RelationBlocks {
			continue
		}

		source := graph.nodes[edge.Source]
		if source.Kind == NodeQuestion && source.Unfinished {
			signals = append(signals, Signal{
				Kind:     SignalBlocked,
				Subject:  edge.Target,
				Related:  edge.Source,
				Relation: edge.Relation,
			})
		}
	}

	return signals
}

func TensionSignals(graph Graph, touched []kernel.RecordID) []Signal {
	inside := make(map[kernel.RecordID]bool, len(touched))
	for _, id := range touched {
		inside[id] = true
	}

	var signals []Signal
	for _, edge := range graph.edges {
		switch {
		case inside[edge.Source] && !inside[edge.Target]:
			signals = append(signals, tension(edge.Source, edge.Target, edge))
		case inside[edge.Target] && !inside[edge.Source]:
			signals = append(signals, tension(edge.Target, edge.Source, edge))
		}
	}

	return signals
}

func CompareSignals(a, b Signal) int {
	return cmp.Or(
		cmp.Compare(a.Kind, b.Kind),
		compareIDs(a.Subject, b.Subject),
		compareIDs(a.Related, b.Related),
		cmp.Compare(a.Relation, b.Relation),
	)
}

func tension(subject, related kernel.RecordID, edge GraphEdge) Signal {
	return Signal{
		Kind:     SignalTension,
		Subject:  subject,
		Related:  related,
		Relation: edge.Relation,
	}
}
