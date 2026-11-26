package handle

import (
	"strings"

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
	"promise_unknown":      "use only promise ids the record already has",
	"graph_invalid":        "fix the references listed in details and rerun",
	"identity_missing":     "set git config user.name, then rerun",
	"actor_invalid":        "set BEACON_ACTOR to agent:<name> or human:<name>",
	"session_invalid":      "pass --session with letters, digits, . _ -",
	"record_changed":       "run the same command again",
	"guide_not_archivable": "delete the guide file; git keeps its history",
	"task_has_open_children": "finish or close the subtasks listed " +
		"in details, then rerun",
	"claim_taken": "task {task} is held by branch {branch}; run " +
		"beacon ready to pick another, or rerun with --force only if " +
		"the person asks to take it",
	"claim_not_task":        "pass a task id; run beacon ready to list them",
	"claim_task_done":       "run beacon ready to pick another task",
	"claim_task_archived":   "run beacon get {task} to see the closed task",
	"claim_branch_detached": "check out a branch, then rerun",
	"base_missing": "pass --base <ref> with the branch to compare " +
		"against",
	"runner_missing": "add runner {runner} to .beacon/config.yaml, or fix " +
		"the check of promise {promise}",
	"capability_missing": "run beacon find <text> to get the capability id",
	"param_missing":      "fix the check params of the promise, then rerun",
	"param_unused":       "fix the check params of the promise, then rerun",
	"param_invalid":      "fix the check params of the promise, then rerun",
	"runner_run_invalid": "fix the runner in .beacon/config.yaml, then rerun",
	"state_invalid":      "fix the reported records, then run beacon doctor",
	"text_invalid":       "pass a text that is not blank",
	"approval_required": "ask the person to approve the changes in details; " +
		"if they say yes, rerun the same command with --approve",
}

var blockedCodes = map[string]bool{
	"approval_required": true,
	"claim_taken":       true,
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
	hint = fillHint(hint, view.Details)

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

func fillHint(hint string, details map[string]any) string {
	var pairs []string

	for key, value := range details {
		if text, isText := value.(string); isText {
			pairs = append(pairs, "{"+key+"}", text)
		}
	}

	return strings.NewReplacer(pairs...).Replace(hint)
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
