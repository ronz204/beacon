package ledger

import (
	"slices"
	"time"

	"github.com/ronz204/beacon/internal/kernel"
)

type GoalState struct {
	ID        GoalID
	Created   kernel.CalendarDate
	Author    kernel.Actor
	Statement kernel.RecordText
	Success   kernel.RecordText
	Nongoals  []kernel.RecordText
	Status    GoalStatus
	Approved  kernel.Approval
}

type GoalInput struct {
	ID        GoalID
	Author    kernel.Actor
	Statement kernel.RecordText
	Success   kernel.RecordText
	Nongoals  []kernel.RecordText
}

type Goal struct {
	state GoalState
}

func NewGoal(input GoalInput, now time.Time) (Goal, error) {
	created, err := calendarDateOf(now)
	if err != nil {
		return Goal{}, err
	}

	state := GoalState{
		ID:        input.ID,
		Created:   created,
		Author:    input.Author,
		Statement: input.Statement,
		Success:   input.Success,
		Nongoals:  input.Nongoals,
		Status:    GoalActive,
	}
	if err := ValidateGoalState(state); err != nil {
		return Goal{}, err
	}

	return RestoreGoal(state), nil
}

func RestoreGoal(state GoalState) Goal {
	state.Nongoals = slices.Clone(state.Nongoals)
	return Goal{state: state}
}

func (g Goal) SameAs(other Goal) bool {
	return g.state.ID == other.state.ID
}

func (g Goal) ID() GoalID {
	return g.state.ID
}

func (g Goal) Created() kernel.CalendarDate {
	return g.state.Created
}

func (g Goal) Author() kernel.Actor {
	return g.state.Author
}

func (g Goal) Statement() kernel.RecordText {
	return g.state.Statement
}

func (g Goal) Success() kernel.RecordText {
	return g.state.Success
}

func (g Goal) Nongoals() []kernel.RecordText {
	return slices.Clone(g.state.Nongoals)
}

func (g Goal) Status() GoalStatus {
	return g.state.Status
}

func (g Goal) Approved() (kernel.Approval, bool) {
	return g.state.Approved, !isMissing(g.state.Approved)
}

func calendarDateOf(now time.Time) (kernel.CalendarDate, error) {
	return kernel.ParseCalendarDate(now.Format(time.DateOnly))
}
