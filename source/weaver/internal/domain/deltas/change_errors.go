package deltas

import "github.com/ronz204/beacon/source/kernel"

type ChangeCode string

const (
	ChangeKindInvalid       ChangeCode = "change_kind_invalid"
	ChangeRecordInvalid     ChangeCode = "change_record_invalid"
	ChangeSourceFailed      ChangeCode = "change_source_failed"
	ChangeTargetInvalid     ChangeCode = "change_target_invalid"
	ChangeRecordKindInvalid ChangeCode = "change_record_kind_invalid"
)

type ChangeError = kernel.Error[ChangeCode]

var changeErrors = kernel.ErrorCatalog[ChangeCode]{
	ChangeKindInvalid:       "change kind must be added, modified or deleted",
	ChangeRecordInvalid:     "changed record is missing a required field",
	ChangeSourceFailed:      "the state changes could not be read",
	ChangeTargetInvalid:     "approval target must name a record or a path",
	ChangeRecordKindInvalid: "record kind must be a record type or config",
}

func invalidChangeValue(code ChangeCode, raw string) ChangeError {
	details := map[string]any{"value": raw}
	return changeErrors.Fail(code, details)
}

func FailChangeSource(cause error) ChangeError {
	return changeErrors.Fail(ChangeSourceFailed, nil).WithCause(cause)
}
