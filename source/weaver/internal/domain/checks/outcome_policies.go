package checks

func OutcomeOf(result ExecutionResult) CheckOutcome {
	switch {
	case result.TimedOut:
		return FailedOutcome(ReasonTimeout)
	case result.ExitCode != 0:
		return FailedOutcome(ReasonExitCode)
	}

	return PassedOutcome()
}
