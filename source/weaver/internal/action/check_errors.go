package action

import (
	"github.com/ronz204/beacon/source/kernel"
)

type CheckCode string

const (
	BaseMissing       CheckCode = "base_missing"
	RunnerMissing     CheckCode = "runner_missing"
	CapabilityMissing CheckCode = "capability_missing"
)

type CheckError = kernel.Error[CheckCode]

var checkErrors = kernel.ErrorCatalog[CheckCode]{
	BaseMissing:       "no base branch was found to compare against",
	RunnerMissing:     "a check names a runner that config.yaml does not have",
	CapabilityMissing: "the id is not a capability",
}

func FailBaseMissing(base string) CheckError {
	return checkErrors.Fail(BaseMissing, map[string]any{"base": base})
}
