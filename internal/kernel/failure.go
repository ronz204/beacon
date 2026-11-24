package kernel

type RecordFailureCode string

const RecordFileInvalid RecordFailureCode = "record_file_invalid"

type RecordFailure = Error[RecordFailureCode]

var recordFailures = ErrorCatalog[RecordFailureCode]{
	RecordFileInvalid: "record file could not be loaded",
}

func FailHistory(cause error) KernelError {
	return kernelErrors.Fail(KernelHistoryFailed, nil).WithCause(cause)
}

func NewRecordFailure(file string, line int, cause error) RecordFailure {
	details := map[string]any{
		"file": file,
		"line": line,
	}

	return recordFailures.Fail(RecordFileInvalid, details).WithCause(cause)
}
