package thread

import "github.com/ronz204/beacon/internal/kernel"

type SignalKind string

const (
	SignalBlocked    SignalKind = "blocked"
	SignalRed        SignalKind = "red"
	SignalUnscoped   SignalKind = "unscoped"
	SignalNotRunHere SignalKind = "not_run_here"
	SignalTension    SignalKind = "tension"
)

type Signal struct {
	Kind     SignalKind
	Subject  kernel.RecordID
	Related  kernel.RecordID
	Relation RelationKind
}
