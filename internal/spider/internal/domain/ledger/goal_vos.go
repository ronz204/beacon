package ledger

import (
	"slices"
	"strings"

	"github.com/ronz204/beacon/internal/kernel"
)

type GoalID struct {
	value kernel.RecordID
}

func ParseGoalID(raw string) (GoalID, error) {
	id, isRoot := parseRootRecordID(raw)

	if !isRoot {
		return GoalID{}, invalidGoalValue(GoalIDInvalid, raw)
	}

	return GoalID{value: id}, nil
}

func (id GoalID) String() string {
	return id.value.String()
}

type GoalStatus string

const (
	GoalActive   GoalStatus = "active"
	GoalAchieved GoalStatus = "achieved"
	GoalDropped  GoalStatus = "dropped"
)

func GoalStatuses() []GoalStatus {
	return []GoalStatus{
		GoalActive,
		GoalAchieved,
		GoalDropped,
	}
}

func ParseGoalStatus(raw string) (GoalStatus, error) {
	status := GoalStatus(strings.TrimSpace(raw))

	if !slices.Contains(GoalStatuses(), status) {
		return "", invalidGoalValue(GoalStatusInvalid, raw)
	}

	return status, nil
}

func parseRootRecordID(raw string) (kernel.RecordID, bool) {
	id, err := kernel.ParseRecordID(raw)
	if err != nil {
		return kernel.RecordID{}, false
	}

	_, hasParent := id.Parent()

	return id, !hasParent
}
