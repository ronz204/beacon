package checks

type OutcomeStatus string

const (
	OutcomePassed OutcomeStatus = "passed"
	OutcomeFailed OutcomeStatus = "failed"
)

type OutcomeReason string

const (
	ReasonExitCode OutcomeReason = "exit_code"
	ReasonTimeout  OutcomeReason = "timeout"
)

type CheckOutcome struct {
	status OutcomeStatus
	reason OutcomeReason
}

func PassedOutcome() CheckOutcome {
	return CheckOutcome{status: OutcomePassed}
}

func FailedOutcome(reason OutcomeReason) CheckOutcome {
	return CheckOutcome{status: OutcomeFailed, reason: reason}
}

func (o CheckOutcome) Status() OutcomeStatus {
	return o.status
}

func (o CheckOutcome) Reason() OutcomeReason {
	return o.reason
}

func (o CheckOutcome) Passed() bool {
	return o.status == OutcomePassed
}
