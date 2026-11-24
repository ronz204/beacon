package ledger

import "github.com/ronz204/beacon/internal/kernel"

type GoalCode string

const (
	GoalIDInvalid     GoalCode = "goal_id_invalid"
	GoalStatusInvalid GoalCode = "goal_status_invalid"
	GoalFieldMissing  GoalCode = "goal_field_missing"
)

type GoalError = kernel.Error[GoalCode]

var goalErrors = kernel.ErrorCatalog[GoalCode]{
	GoalIDInvalid:     "goal id must be a root record id",
	GoalStatusInvalid: "goal status must be active, achieved or dropped",
	GoalFieldMissing:  "goal is missing a required field",
}

func invalidGoalValue(code GoalCode, raw string) GoalError {
	details := map[string]any{"value": raw}
	return goalErrors.Fail(code, details)
}
