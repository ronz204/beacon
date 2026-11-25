package ledger

import (
	"slices"

	"github.com/ronz204/beacon/source/kernel"
)

type PromiseState struct {
	ID       PromiseID
	Text     kernel.RecordText
	Check    PromiseCheck
	Status   PromiseStatus
	Approved kernel.Approval
	Attested PromiseAttestation
	Origin   OriginWorkID
	Refers   []PromiseID
}

type PromiseInput struct {
	ID     PromiseID
	Text   kernel.RecordText
	Check  PromiseCheck
	Origin OriginWorkID
	Refers []PromiseID
}

type Promise struct {
	state PromiseState
}

func (p Promise) SameAs(other Promise) bool {
	return p.state.ID == other.state.ID
}

func (p Promise) ID() PromiseID {
	return p.state.ID
}

func (p Promise) Text() kernel.RecordText {
	return p.state.Text
}

func (p Promise) Check() PromiseCheck {
	return p.state.Check
}

func (p Promise) Status() PromiseStatus {
	return p.state.Status
}

func (p Promise) Approved() (kernel.Approval, bool) {
	return p.state.Approved, !kernel.IsMissing(p.state.Approved)
}

func (p Promise) Attested() (PromiseAttestation, bool) {
	return p.state.Attested, !kernel.IsMissing(p.state.Attested)
}

func (p Promise) Origin() (OriginWorkID, bool) {
	return p.state.Origin, !kernel.IsMissing(p.state.Origin)
}

func (p Promise) Refers() []PromiseID {
	return slices.Clone(p.state.Refers)
}
