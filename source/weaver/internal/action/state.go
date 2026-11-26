package action

import "github.com/ronz204/beacon/source/kernel"

type StateCode string

const (
	StateInvalid StateCode = "state_invalid"
)

type StateError = kernel.Error[StateCode]

var stateErrors = kernel.ErrorCatalog[StateCode]{
	StateInvalid: "the project state has problems",
}

func FailStateInvalid(findings any) StateError {
	return stateErrors.Fail(StateInvalid, map[string]any{
		"findings": findings,
	})
}
