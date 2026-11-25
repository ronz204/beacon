package agenda

import (
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

type WorkID struct {
	value kernel.RecordID
}

func ParseWorkID(raw string) (WorkID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return WorkID{}, invalidWork(WorkIDInvalid, raw)
	}

	return WorkID{value: id}, nil
}

func (id WorkID) String() string {
	return id.value.String()
}

type WorkPromiseID struct {
	value kernel.RecordID
}

func ParseWorkPromiseID(raw string) (WorkPromiseID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return WorkPromiseID{}, invalidWork(WorkPromiseInvalid, raw)
	}

	return WorkPromiseID{value: id}, nil
}

func (id WorkPromiseID) String() string {
	return id.value.String()
}

type WorkGoalID struct {
	value kernel.RecordID
}

func ParseWorkGoalID(raw string) (WorkGoalID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return WorkGoalID{}, invalidWork(WorkGoalInvalid, raw)
	}

	return WorkGoalID{value: id}, nil
}

func (id WorkGoalID) String() string {
	return id.value.String()
}

type WorkRequest struct {
	value string
}

func ParseWorkRequest(raw string) (WorkRequest, error) {
	if strings.TrimSpace(raw) == "" {
		return WorkRequest{}, invalidWork(WorkRequestInvalid, raw)
	}

	return WorkRequest{value: raw}, nil
}

func (r WorkRequest) String() string {
	return r.value
}

type WorkStatus string

const (
	WorkActive  WorkStatus = "active"
	WorkDone    WorkStatus = "done"
	WorkDropped WorkStatus = "dropped"
)

func WorkStatuses() []WorkStatus {
	return []WorkStatus{
		WorkActive,
		WorkDone,
		WorkDropped,
	}
}

func ParseWorkStatus(raw string) (WorkStatus, error) {
	status := WorkStatus(strings.TrimSpace(raw))

	if !slices.Contains(WorkStatuses(), status) {
		return "", invalidWork(WorkStatusInvalid, raw)
	}

	return status, nil
}
