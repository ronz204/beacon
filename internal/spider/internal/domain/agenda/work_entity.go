package agenda

import (
	"slices"
	"time"

	"github.com/ronz204/beacon/internal/kernel"
)

type WorkState struct {
	ID         WorkID
	Created    kernel.CalendarDate
	Author     kernel.Actor
	Request    WorkRequest
	Implements []WorkPromiseID
	Serves     WorkGoalID
	Log        []kernel.RecordText
	Status     WorkStatus
	Approved   kernel.Approval
}

type WorkInput struct {
	ID         WorkID
	Author     kernel.Actor
	Request    WorkRequest
	Implements []WorkPromiseID
	Serves     WorkGoalID
}

type Work struct {
	state WorkState
}

func NewWork(input WorkInput, now time.Time) (Work, error) {
	created, err := kernel.CalendarDateOf(now)
	if err != nil {
		return Work{}, err
	}

	state := WorkState{
		ID:         input.ID,
		Created:    created,
		Author:     input.Author,
		Request:    input.Request,
		Implements: input.Implements,
		Serves:     input.Serves,
		Status:     WorkActive,
	}
	if err := ValidateWorkState(state); err != nil {
		return Work{}, err
	}

	return RestoreWork(state), nil
}

func RestoreWork(state WorkState) Work {
	state.Implements = slices.Clone(state.Implements)
	state.Log = slices.Clone(state.Log)

	return Work{state: state}
}

func (w Work) SameAs(other Work) bool {
	return w.state.ID == other.state.ID
}

func (w Work) ID() WorkID {
	return w.state.ID
}

func (w Work) Created() kernel.CalendarDate {
	return w.state.Created
}

func (w Work) Author() kernel.Actor {
	return w.state.Author
}

func (w Work) Request() WorkRequest {
	return w.state.Request
}

func (w Work) Implements() []WorkPromiseID {
	return slices.Clone(w.state.Implements)
}

func (w Work) Serves() (WorkGoalID, bool) {
	return w.state.Serves, !kernel.IsMissing(w.state.Serves)
}

func (w Work) Log() []kernel.RecordText {
	return slices.Clone(w.state.Log)
}

func (w Work) Status() WorkStatus {
	return w.state.Status
}

func (w Work) Approved() (kernel.Approval, bool) {
	return w.state.Approved, !kernel.IsMissing(w.state.Approved)
}
