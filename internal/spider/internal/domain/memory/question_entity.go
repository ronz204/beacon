package memory

import (
	"slices"
	"time"

	"github.com/ronz204/beacon/internal/kernel"
)

type QuestionState struct {
	ID       QuestionID
	Created  kernel.CalendarDate
	Author   kernel.Actor
	Text     kernel.RecordText
	Blocks   []QuestionBlockID
	Status   QuestionStatus
	Answer   kernel.RecordText
	Origin   QuestionOriginID
	Approved kernel.Approval
}

type QuestionInput struct {
	ID     QuestionID
	Author kernel.Actor
	Text   kernel.RecordText
	Blocks []QuestionBlockID
	Origin QuestionOriginID
}

type Question struct {
	state QuestionState
}

func NewQuestion(input QuestionInput, now time.Time) (Question, error) {
	created, err := kernel.CalendarDateOf(now)
	if err != nil {
		return Question{}, err
	}

	state := QuestionState{
		ID:      input.ID,
		Created: created,
		Author:  input.Author,
		Text:    input.Text,
		Blocks:  input.Blocks,
		Status:  QuestionOpen,
		Origin:  input.Origin,
	}
	if err := ValidateQuestionState(state); err != nil {
		return Question{}, err
	}

	return RestoreQuestion(state), nil
}

func RestoreQuestion(state QuestionState) Question {
	state.Blocks = slices.Clone(state.Blocks)
	return Question{state: state}
}

func (q Question) SameAs(other Question) bool {
	return q.state.ID == other.state.ID
}

func (q Question) ID() QuestionID {
	return q.state.ID
}

func (q Question) Created() kernel.CalendarDate {
	return q.state.Created
}

func (q Question) Author() kernel.Actor {
	return q.state.Author
}

func (q Question) Text() kernel.RecordText {
	return q.state.Text
}

func (q Question) Blocks() []QuestionBlockID {
	return slices.Clone(q.state.Blocks)
}

func (q Question) Status() QuestionStatus {
	return q.state.Status
}

func (q Question) Answer() (kernel.RecordText, bool) {
	return q.state.Answer, !kernel.IsMissing(q.state.Answer)
}

func (q Question) Origin() (QuestionOriginID, bool) {
	return q.state.Origin, !kernel.IsMissing(q.state.Origin)
}

func (q Question) Approved() (kernel.Approval, bool) {
	return q.state.Approved, !kernel.IsMissing(q.state.Approved)
}
