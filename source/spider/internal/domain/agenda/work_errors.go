package agenda

import "github.com/ronz204/beacon/source/kernel"

type WorkCode string

const (
	WorkIDInvalid      WorkCode = "work_id_invalid"
	WorkRequestInvalid WorkCode = "work_request_invalid"
	WorkStatusInvalid  WorkCode = "work_status_invalid"
	WorkPromiseInvalid WorkCode = "work_promise_invalid"
	WorkGoalInvalid    WorkCode = "work_goal_invalid"
	WorkFieldMissing   WorkCode = "work_field_missing"
)

type WorkError = kernel.Error[WorkCode]

var workErrors = kernel.ErrorCatalog[WorkCode]{
	WorkIDInvalid:      "work id must be a root record id",
	WorkRequestInvalid: "work request must not be blank",
	WorkStatusInvalid:  "work status must be active, done or dropped",
	WorkPromiseInvalid: "work implements must be root promise ids",
	WorkGoalInvalid:    "work serves must be a root goal id",
	WorkFieldMissing:   "work is missing a required field",
}

func invalidWork(code WorkCode, raw string) WorkError {
	details := map[string]any{"value": raw}
	return workErrors.Fail(code, details)
}
