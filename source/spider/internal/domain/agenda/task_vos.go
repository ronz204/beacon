package agenda

import (
	"slices"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/source/kernel"
)

type TaskID struct {
	value kernel.RecordID
}

func ParseTaskID(raw string) (TaskID, error) {
	id, err := kernel.ParseRecordID(raw)
	if err != nil {
		return TaskID{}, invalidTask(TaskIDInvalid, raw)
	}

	if _, hasParent := id.Parent(); !hasParent {
		return TaskID{}, invalidTask(TaskIDInvalid, raw)
	}

	return TaskID{value: id}, nil
}

func (id TaskID) String() string {
	return id.value.String()
}

type TaskWorkID struct {
	value kernel.RecordID
}

func ParseTaskWorkID(raw string) (TaskWorkID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return TaskWorkID{}, invalidTask(TaskParentInvalid, raw)
	}

	return TaskWorkID{value: id}, nil
}

func (id TaskWorkID) String() string {
	return id.value.String()
}

type TaskPromiseID struct {
	value kernel.RecordID
}

func ParseTaskPromiseID(raw string) (TaskPromiseID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return TaskPromiseID{}, invalidTask(TaskPromiseInvalid, raw)
	}

	return TaskPromiseID{value: id}, nil
}

func (id TaskPromiseID) String() string {
	return id.value.String()
}

type TaskBranch struct {
	value string
}

func ParseTaskBranch(raw string) (TaskBranch, error) {
	value := strings.TrimSpace(raw)

	if !isTaskBranch(value) {
		return TaskBranch{}, invalidTask(TaskBranchInvalid, raw)
	}

	return TaskBranch{value: value}, nil
}

func (b TaskBranch) String() string {
	return b.value
}

func isTaskBranch(value string) bool {
	if value == "" || strings.HasPrefix(value, "-") {
		return false
	}

	return !strings.ContainsFunc(value, isBranchBreak)
}

func isBranchBreak(char rune) bool {
	return unicode.IsSpace(char) || unicode.IsControl(char)
}

type TaskStatus string

const (
	TaskOpen    TaskStatus = "open"
	TaskClaimed TaskStatus = "claimed"
	TaskDone    TaskStatus = "done"
)

func TaskStatuses() []TaskStatus {
	return []TaskStatus{
		TaskOpen,
		TaskClaimed,
		TaskDone,
	}
}

func ParseTaskStatus(raw string) (TaskStatus, error) {
	status := TaskStatus(strings.TrimSpace(raw))

	if !slices.Contains(TaskStatuses(), status) {
		return "", invalidTask(TaskStatusInvalid, raw)
	}

	return status, nil
}
