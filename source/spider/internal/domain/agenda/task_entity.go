package agenda

import (
	"slices"
	"time"

	"github.com/ronz204/beacon/source/kernel"
)

type TaskState struct {
	ID         TaskID
	Created    kernel.CalendarDate
	Author     kernel.Actor
	Title      kernel.RecordText
	Parent     TaskWorkID
	Blocks     []TaskID
	Implements TaskPromiseID
	Owner      kernel.Actor
	Branch     TaskBranch
	Status     TaskStatus
	Approved   kernel.Approval
}

type TaskInput struct {
	ID         TaskID
	Author     kernel.Actor
	Title      kernel.RecordText
	Parent     TaskWorkID
	Blocks     []TaskID
	Implements TaskPromiseID
}

type Task struct {
	state TaskState
}

func NewTask(input TaskInput, now time.Time) (Task, error) {
	created, err := kernel.CalendarDateOf(now)
	if err != nil {
		return Task{}, err
	}

	state := TaskState{
		ID:         input.ID,
		Created:    created,
		Author:     input.Author,
		Title:      input.Title,
		Parent:     input.Parent,
		Blocks:     input.Blocks,
		Implements: input.Implements,
		Status:     TaskOpen,
	}
	if err := ValidateTaskState(state); err != nil {
		return Task{}, err
	}

	return RestoreTask(state), nil
}

func RestoreTask(state TaskState) Task {
	state.Blocks = slices.Clone(state.Blocks)
	return Task{state: state}
}

func (t Task) SameAs(other Task) bool {
	return t.state.ID == other.state.ID
}

func (t Task) ID() TaskID {
	return t.state.ID
}

func (t Task) Created() kernel.CalendarDate {
	return t.state.Created
}

func (t Task) Author() kernel.Actor {
	return t.state.Author
}

func (t Task) Title() kernel.RecordText {
	return t.state.Title
}

func (t Task) Parent() TaskWorkID {
	return t.state.Parent
}

func (t Task) Blocks() []TaskID {
	return slices.Clone(t.state.Blocks)
}

func (t Task) Implements() (TaskPromiseID, bool) {
	return t.state.Implements, !kernel.IsMissing(t.state.Implements)
}

func (t Task) Owner() (kernel.Actor, bool) {
	return t.state.Owner, !kernel.IsMissing(t.state.Owner)
}

func (t Task) Branch() (TaskBranch, bool) {
	return t.state.Branch, !kernel.IsMissing(t.state.Branch)
}

func (t Task) Status() TaskStatus {
	return t.state.Status
}

func (t Task) Approved() (kernel.Approval, bool) {
	return t.state.Approved, !kernel.IsMissing(t.state.Approved)
}
