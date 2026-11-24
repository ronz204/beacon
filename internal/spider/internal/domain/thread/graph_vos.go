package thread

import "github.com/ronz204/beacon/internal/kernel"

type NodeKind string

const (
	NodeGoal       NodeKind = "goal"
	NodeCapability NodeKind = "capability"
	NodePromise    NodeKind = "promise"
	NodeDecision   NodeKind = "decision"
	NodeQuestion   NodeKind = "question"
	NodeGuide      NodeKind = "guide"
	NodeWork       NodeKind = "work"
	NodeTask       NodeKind = "task"
)

type RelationKind string

const (
	RelationImplements RelationKind = "implements"
	RelationBlocks     RelationKind = "blocks"
	RelationParent     RelationKind = "parent"
	RelationSupersedes RelationKind = "supersedes"
	RelationServes     RelationKind = "serves"
	RelationOrigin     RelationKind = "origin"
	RelationRefers     RelationKind = "refers"
)

type GraphNode struct {
	ID         kernel.RecordID
	Kind       NodeKind
	Owner      kernel.RecordID
	Unfinished bool
}

type GraphEdge struct {
	Source   kernel.RecordID
	Relation RelationKind
	Target   kernel.RecordID
}
