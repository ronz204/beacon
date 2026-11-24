package ledger

import (
	"slices"
	"time"

	"github.com/ronz204/beacon/internal/kernel"
)

type CapabilityState struct {
	ID       CapabilityID
	Created  kernel.CalendarDate
	Author   kernel.Actor
	Name     kernel.RecordText
	Intent   kernel.RecordText
	Paths    []kernel.PathGlob
	Serves   GoalID
	Nongoals []kernel.RecordText
	Promises []PromiseState
	Approved kernel.Approval
}

type CapabilityInput struct {
	ID       CapabilityID
	Author   kernel.Actor
	Name     kernel.RecordText
	Intent   kernel.RecordText
	Paths    []kernel.PathGlob
	Serves   GoalID
	Nongoals []kernel.RecordText
}

type Capability struct {
	state CapabilityState
}

func NewCapability(
	input CapabilityInput,
	now time.Time,
) (Capability, error) {
	created, err := kernel.CalendarDateOf(now)
	if err != nil {
		return Capability{}, err
	}

	state := CapabilityState{
		ID:       input.ID,
		Created:  created,
		Author:   input.Author,
		Name:     input.Name,
		Intent:   input.Intent,
		Paths:    input.Paths,
		Serves:   input.Serves,
		Nongoals: input.Nongoals,
	}
	if err := ValidateCapabilityState(state); err != nil {
		return Capability{}, err
	}

	return RestoreCapability(state), nil
}

func RestoreCapability(state CapabilityState) Capability {
	return Capability{state: cloneCapabilityState(state)}
}

func (c Capability) ProposePromise(input PromiseInput) (Capability, error) {
	next := cloneCapabilityState(c.state)
	next.Promises = append(next.Promises, PromiseState{
		ID:     input.ID,
		Text:   input.Text,
		Check:  input.Check,
		Status: PromiseProposed,
		Origin: input.Origin,
		Refers: slices.Clone(input.Refers),
	})

	if err := ValidateCapabilityState(next); err != nil {
		return Capability{}, err
	}

	return Capability{state: next}, nil
}

func (c Capability) SameAs(other Capability) bool {
	return c.state.ID == other.state.ID
}

func (c Capability) ID() CapabilityID {
	return c.state.ID
}

func (c Capability) Created() kernel.CalendarDate {
	return c.state.Created
}

func (c Capability) Author() kernel.Actor {
	return c.state.Author
}

func (c Capability) Name() kernel.RecordText {
	return c.state.Name
}

func (c Capability) Intent() kernel.RecordText {
	return c.state.Intent
}

func (c Capability) Paths() []kernel.PathGlob {
	return slices.Clone(c.state.Paths)
}

func (c Capability) Serves() (GoalID, bool) {
	return c.state.Serves, !kernel.IsMissing(c.state.Serves)
}

func (c Capability) Nongoals() []kernel.RecordText {
	return slices.Clone(c.state.Nongoals)
}

func (c Capability) Promises() []Promise {
	promises := make([]Promise, 0, len(c.state.Promises))

	for _, state := range c.state.Promises {
		promises = append(promises, Promise{state: state})
	}

	return promises
}

func (c Capability) Approved() (kernel.Approval, bool) {
	return c.state.Approved, !kernel.IsMissing(c.state.Approved)
}

func cloneCapabilityState(state CapabilityState) CapabilityState {
	state.Paths = slices.Clone(state.Paths)
	state.Nongoals = slices.Clone(state.Nongoals)
	state.Promises = slices.Clone(state.Promises)

	for index := range state.Promises {
		refers := slices.Clone(state.Promises[index].Refers)
		state.Promises[index].Refers = refers
	}

	return state
}
