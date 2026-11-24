package memory

import (
	"slices"
	"time"

	"github.com/ronz204/beacon/internal/kernel"
)

type DecisionState struct {
	ID         DecisionID
	Created    kernel.CalendarDate
	Author     kernel.Actor
	What       kernel.RecordText
	Why        kernel.RecordText
	Where      []kernel.PathGlob
	Supersedes DecisionID
	Rejected   []kernel.RecordText
	Status     DecisionStatus
	Review     ReviewSpan
	Reviewed   kernel.CalendarDate
	Approved   kernel.Approval
}

type DecisionInput struct {
	ID         DecisionID
	Author     kernel.Actor
	What       kernel.RecordText
	Why        kernel.RecordText
	Where      []kernel.PathGlob
	Supersedes DecisionID
	Rejected   []kernel.RecordText
	Review     ReviewSpan
}

type Decision struct {
	state DecisionState
}

func NewDecision(input DecisionInput, now time.Time) (Decision, error) {
	created, err := calendarDateOf(now)
	if err != nil {
		return Decision{}, err
	}

	state := DecisionState{
		ID:         input.ID,
		Created:    created,
		Author:     input.Author,
		What:       input.What,
		Why:        input.Why,
		Where:      input.Where,
		Supersedes: input.Supersedes,
		Rejected:   input.Rejected,
		Status:     DecisionActive,
		Review:     input.Review,
	}
	if err := ValidateDecisionState(state); err != nil {
		return Decision{}, err
	}

	return RestoreDecision(state), nil
}

func RestoreDecision(state DecisionState) Decision {
	state.Where = slices.Clone(state.Where)
	state.Rejected = slices.Clone(state.Rejected)

	return Decision{state: state}
}

func (d Decision) SameAs(other Decision) bool {
	return d.state.ID == other.state.ID
}

func (d Decision) ID() DecisionID {
	return d.state.ID
}

func (d Decision) Created() kernel.CalendarDate {
	return d.state.Created
}

func (d Decision) Author() kernel.Actor {
	return d.state.Author
}

func (d Decision) What() kernel.RecordText {
	return d.state.What
}

func (d Decision) Why() kernel.RecordText {
	return d.state.Why
}

func (d Decision) Where() []kernel.PathGlob {
	return slices.Clone(d.state.Where)
}

func (d Decision) Supersedes() (DecisionID, bool) {
	return d.state.Supersedes, !isMissing(d.state.Supersedes)
}

func (d Decision) Rejected() []kernel.RecordText {
	return slices.Clone(d.state.Rejected)
}

func (d Decision) Status() DecisionStatus {
	return d.state.Status
}

func (d Decision) Review() (ReviewSpan, bool) {
	return d.state.Review, !isMissing(d.state.Review)
}

func (d Decision) Reviewed() (kernel.CalendarDate, bool) {
	return d.state.Reviewed, !isMissing(d.state.Reviewed)
}

func (d Decision) Approved() (kernel.Approval, bool) {
	return d.state.Approved, !isMissing(d.state.Approved)
}

func calendarDateOf(now time.Time) (kernel.CalendarDate, error) {
	return kernel.ParseCalendarDate(now.Format(time.DateOnly))
}
