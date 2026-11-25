package agenda

import "github.com/ronz204/beacon/source/kernel"

type TaskCode string

const (
	TaskIDInvalid      TaskCode = "task_id_invalid"
	TaskStatusInvalid  TaskCode = "task_status_invalid"
	TaskParentInvalid  TaskCode = "task_parent_invalid"
	TaskParentMismatch TaskCode = "task_parent_mismatch"
	TaskPromiseInvalid TaskCode = "task_promise_invalid"
	TaskBranchInvalid  TaskCode = "task_branch_invalid"
	TaskOwnerMissing   TaskCode = "task_owner_missing"
	TaskFieldMissing   TaskCode = "task_field_missing"
)

type TaskError = kernel.Error[TaskCode]

var taskErrors = kernel.ErrorCatalog[TaskCode]{
	TaskIDInvalid:      "task id must be a child id like bc-91cd.1",
	TaskStatusInvalid:  "task status must be open, claimed or done",
	TaskParentInvalid:  "task parent must be a root work id",
	TaskParentMismatch: "task id must descend from its parent work",
	TaskPromiseInvalid: "task implements must be a root promise id",
	TaskBranchInvalid:  "task branch must be one safe word",
	TaskOwnerMissing:   "a claimed task needs an owner",
	TaskFieldMissing:   "task is missing a required field",
}

func invalidTask(code TaskCode, raw string) TaskError {
	details := map[string]any{"value": raw}
	return taskErrors.Fail(code, details)
}
