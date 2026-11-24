package memory

import (
	"slices"

	"github.com/ronz204/beacon/internal/kernel"
)

type GuideState struct {
	ID       GuideID
	Title    kernel.RecordText
	Kind     GuideKind
	Applies  []kernel.PathGlob
	Always   bool
	Review   ReviewSpan
	Reviewed kernel.CalendarDate
	Tags     []GuideTag
	Audience GuideAudience
	Status   GuideStatus
	Refers   []GuideCapabilityID
	Lines    int
}

type Guide struct {
	state GuideState
}

func NewGuide(input GuideState) (Guide, error) {
	if err := ValidateGuideState(input); err != nil {
		return Guide{}, err
	}

	return RestoreGuide(input), nil
}

func RestoreGuide(state GuideState) Guide {
	state.Applies = slices.Clone(state.Applies)
	state.Tags = slices.Clone(state.Tags)
	state.Refers = slices.Clone(state.Refers)

	return Guide{state: state}
}

func (g Guide) SameAs(other Guide) bool {
	return g.state.ID == other.state.ID
}

func (g Guide) ID() GuideID {
	return g.state.ID
}

func (g Guide) Title() kernel.RecordText {
	return g.state.Title
}

func (g Guide) Kind() GuideKind {
	return g.state.Kind
}

func (g Guide) Applies() []kernel.PathGlob {
	return slices.Clone(g.state.Applies)
}

func (g Guide) Always() bool {
	return g.state.Always
}

func (g Guide) Review() (ReviewSpan, bool) {
	return g.state.Review, !kernel.IsMissing(g.state.Review)
}

func (g Guide) Reviewed() (kernel.CalendarDate, bool) {
	return g.state.Reviewed, !kernel.IsMissing(g.state.Reviewed)
}

func (g Guide) Tags() []GuideTag {
	return slices.Clone(g.state.Tags)
}

func (g Guide) Audience() (GuideAudience, bool) {
	return g.state.Audience, !kernel.IsMissing(g.state.Audience)
}

func (g Guide) Status() (GuideStatus, bool) {
	return g.state.Status, !kernel.IsMissing(g.state.Status)
}

func (g Guide) Refers() []GuideCapabilityID {
	return slices.Clone(g.state.Refers)
}

func (g Guide) Lines() int {
	return g.state.Lines
}
