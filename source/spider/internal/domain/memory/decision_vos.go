package memory

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

var reviewSpanPattern = regexp.MustCompile(`^[1-9][0-9]*(mo|d)$`)

type DecisionID struct {
	value kernel.RecordID
}

func ParseDecisionID(raw string) (DecisionID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return DecisionID{}, invalidDecision(DecisionIDInvalid, raw)
	}

	return DecisionID{value: id}, nil
}

func (id DecisionID) String() string {
	return id.value.String()
}

type DecisionStatus string

const (
	DecisionActive     DecisionStatus = "active"
	DecisionSuperseded DecisionStatus = "superseded"
)

func DecisionStatuses() []DecisionStatus {
	return []DecisionStatus{
		DecisionActive,
		DecisionSuperseded,
	}
}

func ParseDecisionStatus(raw string) (DecisionStatus, error) {
	status := DecisionStatus(strings.TrimSpace(raw))

	if !slices.Contains(DecisionStatuses(), status) {
		return "", invalidDecision(DecisionStatusInvalid, raw)
	}

	return status, nil
}

type ReviewSpan struct {
	value string
}

func ParseReviewSpan(raw string) (ReviewSpan, error) {
	value := strings.TrimSpace(raw)

	if !reviewSpanPattern.MatchString(value) {
		return ReviewSpan{}, invalidDecision(ReviewSpanInvalid, raw)
	}

	return ReviewSpan{value: value}, nil
}

func (s ReviewSpan) DueAfter(base kernel.CalendarDate) kernel.CalendarDate {
	if amount, isMonths := strings.CutSuffix(s.value, "mo"); isMonths {
		months, _ := strconv.Atoi(amount)
		return base.Plus(months, 0)
	}

	days, _ := strconv.Atoi(strings.TrimSuffix(s.value, "d"))

	return base.Plus(0, days)
}

func (s ReviewSpan) String() string {
	return s.value
}
