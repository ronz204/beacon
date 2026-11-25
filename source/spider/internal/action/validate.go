package action

import (
	"fmt"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

type graphBuilder struct {
	nodes    []thread.GraphNode
	edges    []thread.GraphEdge
	failures []error
}

func Validate(snapshot Snapshot) (thread.Graph, []error) {
	builder := &graphBuilder{}

	for _, goal := range snapshot.Goals {
		builder.addNode(goal.ID(), thread.NodeGoal)
	}
	for _, capability := range snapshot.Capabilities {
		builder.addCapability(capability)
	}
	for _, decision := range snapshot.Decisions {
		builder.addDecision(decision)
	}
	for _, question := range snapshot.Questions {
		builder.addQuestion(question)
	}
	for _, guide := range snapshot.Guides {
		builder.addNode(guide.ID(), thread.NodeGuide)
		addEdges(builder, guide.ID(), thread.RelationRefers, guide.Refers())
	}
	for _, work := range snapshot.Works {
		builder.addWork(work)
	}
	for _, task := range snapshot.Tasks {
		builder.addTask(task)
	}
	builder.nodes = append(builder.nodes, snapshot.Archived...)

	graph, failures := thread.NewGraph(builder.nodes, builder.edges)

	return graph, append(builder.failures, failures...)
}

func (b *graphBuilder) recordID(id fmt.Stringer) (kernel.RecordID, bool) {
	parsed, err := recordIDOf(id)
	if err != nil {
		b.failures = append(b.failures, err)
		return kernel.RecordID{}, false
	}

	return parsed, true
}

func recordIDOf(id fmt.Stringer) (kernel.RecordID, error) {
	return kernel.ParseRecordID(id.String())
}

func (b *graphBuilder) addNode(id fmt.Stringer, kind thread.NodeKind) {
	if parsed, valid := b.recordID(id); valid {
		b.nodes = append(b.nodes, thread.GraphNode{ID: parsed, Kind: kind})
	}
}

func (b *graphBuilder) addEdge(
	source fmt.Stringer,
	relation thread.RelationKind,
	target fmt.Stringer,
) {
	parsedSource, validSource := b.recordID(source)
	parsedTarget, validTarget := b.recordID(target)

	if validSource && validTarget {
		b.edges = append(b.edges, thread.GraphEdge{
			Source:   parsedSource,
			Relation: relation,
			Target:   parsedTarget,
		})
	}
}

func addEdges[T fmt.Stringer](
	b *graphBuilder,
	source fmt.Stringer,
	relation thread.RelationKind,
	targets []T,
) {
	for _, target := range targets {
		b.addEdge(source, relation, target)
	}
}

func (b *graphBuilder) addCapability(capability ledger.Capability) {
	b.addNode(capability.ID(), thread.NodeCapability)

	if goal, found := capability.Serves(); found {
		b.addEdge(capability.ID(), thread.RelationServes, goal)
	}

	owner, valid := b.recordID(capability.ID())
	if !valid {
		return
	}

	for _, promise := range capability.Promises() {
		b.addPromise(owner, promise)
	}
}

func (b *graphBuilder) addPromise(
	owner kernel.RecordID,
	promise ledger.Promise,
) {
	if id, valid := b.recordID(promise.ID()); valid {
		b.nodes = append(b.nodes, thread.GraphNode{
			ID:    id,
			Kind:  thread.NodePromise,
			Owner: owner,
		})
	}

	if work, found := promise.Origin(); found {
		b.addEdge(promise.ID(), thread.RelationOrigin, work)
	}

	addEdges(b, promise.ID(), thread.RelationRefers, promise.Refers())
}

func (b *graphBuilder) addDecision(decision memory.Decision) {
	b.addNode(decision.ID(), thread.NodeDecision)

	if replaced, found := decision.Supersedes(); found {
		b.addEdge(decision.ID(), thread.RelationSupersedes, replaced)
	}
}

func (b *graphBuilder) addQuestion(question memory.Question) {
	if id, valid := b.recordID(question.ID()); valid {
		b.nodes = append(b.nodes, thread.GraphNode{
			ID:         id,
			Kind:       thread.NodeQuestion,
			Unfinished: question.Status() == memory.QuestionOpen,
		})
	}

	addEdges(b, question.ID(), thread.RelationBlocks, question.Blocks())

	if work, found := question.Origin(); found {
		b.addEdge(question.ID(), thread.RelationOrigin, work)
	}
}

func (b *graphBuilder) addWork(work agenda.Work) {
	b.addNode(work.ID(), thread.NodeWork)
	addEdges(b, work.ID(), thread.RelationImplements, work.Implements())

	if goal, found := work.Serves(); found {
		b.addEdge(work.ID(), thread.RelationServes, goal)
	}
}

func (b *graphBuilder) addTask(task agenda.Task) {
	if id, valid := b.recordID(task.ID()); valid {
		b.nodes = append(b.nodes, thread.GraphNode{
			ID:         id,
			Kind:       thread.NodeTask,
			Unfinished: task.Status() != agenda.TaskDone,
		})
	}

	b.addEdge(task.ID(), thread.RelationParent, task.Parent())
	addEdges(b, task.ID(), thread.RelationBlocks, task.Blocks())

	if promise, found := task.Implements(); found {
		b.addEdge(task.ID(), thread.RelationImplements, promise)
	}
}
