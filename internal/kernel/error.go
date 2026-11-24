package kernel

type Error[C ~string] struct {
	Code    C
	Message string
	Details map[string]any
	Cause   error
}

func (e Error[C]) Error() string {
	return string(e.Code) + ": " + e.Message
}

func (e Error[C]) Unwrap() error {
	return e.Cause
}

func (e Error[C]) WithCause(cause error) Error[C] {
	e.Cause = cause
	return e
}

type ErrorCatalog[C ~string] map[C]string

func (c ErrorCatalog[C]) Fail(
	code C,
	details map[string]any,
) Error[C] {
	return Error[C]{
		Code:    code,
		Message: c[code],
		Details: details,
	}
}

type KernelCode string

const (
	KernelActorInvalid    KernelCode = "actor_invalid"
	KernelApprovalInvalid KernelCode = "approval_invalid"
	KernelDateInvalid     KernelCode = "date_invalid"
	KernelGlobInvalid     KernelCode = "glob_invalid"
	KernelIDIssueFailed   KernelCode = "id_issue_failed"
	KernelRecordIDInvalid KernelCode = "record_id_invalid"
	KernelTextInvalid     KernelCode = "text_invalid"
)

type KernelError = Error[KernelCode]

var kernelErrors = ErrorCatalog[KernelCode]{
	KernelActorInvalid:    "actor must be human:<name> or agent:<name>",
	KernelApprovalInvalid: "approval via must be an agent",
	KernelDateInvalid:     "date must be a real day in YYYY-MM-DD format",
	KernelGlobInvalid:     "glob must be repo-relative using *, **, ?, {a,b}",
	KernelIDIssueFailed:   "a new record id could not be issued",
	KernelRecordIDInvalid: "record id must look like bc-a1b2 or bc-a1b2.1",
	KernelTextInvalid:     "text must not be blank",
}

func invalidValue(code KernelCode, raw string) KernelError {
	details := map[string]any{"value": raw}
	return kernelErrors.Fail(code, details)
}
