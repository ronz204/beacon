package thread_test

import (
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

func question(raw string, open bool) thread.GraphNode {
	return thread.GraphNode{
		ID:         recordID(raw),
		Kind:       thread.NodeQuestion,
		Unfinished: open,
	}
}

func signalGraph(t *testing.T, open bool) thread.Graph {
	t.Helper()

	nodes := []thread.GraphNode{
		node(workID, thread.NodeWork),
		task(taskID, true),
		promise(promiseID, capabilityID),
		node(capabilityID, thread.NodeCapability),
		question(questionID, open),
	}
	edges := []thread.GraphEdge{
		edge(questionID, relBlocks, workID),
		edge(workID, relImplements, promiseID),
		edge(taskID, relParent, workID),
	}

	graph, failures := thread.NewGraph(nodes, edges)
	requireCodes(t, failures)

	return graph
}

func TestBlockedSignalsComeFromOpenQuestions(t *testing.T) {
	signals := thread.BlockedSignals(signalGraph(t, true))

	if len(signals) != 1 {
		t.Fatalf("signals = %+v", signals)
	}

	signal := signals[0]
	if signal.Kind != thread.SignalBlocked {
		t.Fatalf("kind = %q", signal.Kind)
	}
	if signal.Subject != recordID(workID) {
		t.Fatalf("subject = %q", signal.Subject)
	}
	if signal.Related != recordID(questionID) {
		t.Fatalf("related = %q", signal.Related)
	}
}

func TestAnsweredQuestionsDoNotBlock(t *testing.T) {
	signals := thread.BlockedSignals(signalGraph(t, false))

	if len(signals) != 0 {
		t.Fatalf("signals = %+v", signals)
	}
}

func TestTensionSignalsPointOutsideTheTouchedSet(t *testing.T) {
	graph := signalGraph(t, true)
	touched := []kernel.RecordID{recordID(promiseID)}

	signals := thread.TensionSignals(graph, touched)

	if len(signals) != 1 {
		t.Fatalf("signals = %+v", signals)
	}
	if signals[0].Kind != thread.SignalTension {
		t.Fatalf("kind = %q", signals[0].Kind)
	}
	if signals[0].Related != recordID(workID) {
		t.Fatalf("related = %q", signals[0].Related)
	}
	if signals[0].Relation != relImplements {
		t.Fatalf("relation = %q", signals[0].Relation)
	}
}

func TestTensionSignalsIgnoreLinksInsideTheTouchedSet(t *testing.T) {
	graph := signalGraph(t, true)
	touched := []kernel.RecordID{
		recordID(promiseID),
		recordID(workID),
		recordID(questionID),
		recordID(taskID),
	}

	if signals := thread.TensionSignals(graph, touched); len(signals) != 0 {
		t.Fatalf("signals = %+v", signals)
	}
}

func TestGraphNodeLookup(t *testing.T) {
	graph := signalGraph(t, true)

	found, exists := graph.Node(recordID(questionID))
	if !exists || found.Kind != thread.NodeQuestion {
		t.Fatalf("node = %+v, exists = %v", found, exists)
	}

	if _, exists := graph.Node(recordID("bc-ffff")); exists {
		t.Fatal("unknown node should not exist")
	}
}
