package handle

import (
	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/router/internal/output"
	"github.com/ronz204/beacon/source/spider"
)

const (
	stateInvalidCode    = "state_invalid"
	stateInvalidMessage = "the project state has problems"
	commandFailedHint   = "run beacon doctor"
)

var errorHints = map[string]string{
	"record_missing":  "run beacon find <text> to get the id",
	"payload_invalid": "fix the fields listed in details and rerun",
	"promise_omitted": "keep the promise, or run " +
		"beacon close <promise-id> --reason <text>",
	"promise_unknown":  "use only promise ids the record already has",
	"graph_invalid":    "fix the references listed in details and rerun",
	"identity_missing": "set git config user.name, then rerun",
	"actor_invalid":    "set BEACON_ACTOR to agent:<name> or human:<name>",
	"session_invalid":  "pass --session with letters, digits, . _ -",
	"record_changed":   "run the same command again",
	"approval_required": "ask the person to approve the changes in details; " +
		"if they say yes, rerun the same command with --approve",
}

var blockedCodes = map[string]bool{
	"approval_required": true,
}

func errorFailure(err error) output.Failure {
	view := kernel.DescribeFailure(err)

	code := view.Code
	if code == "" {
		code = failureFallback
	}

	hint, known := errorHints[code]
	if !known {
		hint = commandFailedHint
	}

	message := view.Message
	if message == "" {
		message = view.Cause
	}

	failure := output.Failure{
		Code:    code,
		Message: message,
		Hint:    hint,
		Details: view.Details,
		Lines:   failureLines(view),
	}
	if blockedCodes[code] {
		failure.Exit = output.ExitBlocked
	}

	return failure
}

func failureLines(view kernel.FailureView) []string {
	if view.File == "" {
		return nil
	}

	return []string{view.File}
}

func stateFailure(findings []spider.Finding) *output.Failure {
	if len(errorFindings(findings)) == 0 {
		return nil
	}

	return &output.Failure{
		Code:    stateInvalidCode,
		Message: stateInvalidMessage,
		Hint:    findingsHint(findings),
		Details: doctorDetails{Findings: findings},
		Lines:   findingLines(findings),
	}
}
