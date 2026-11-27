package record

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
)

var guideFilePattern = regexp.MustCompile(
	`^(bc-[0-9a-f]{4,})-.+\.md$`,
)

func decodeGoal(file string, data []byte) (ledger.Goal, []error) {
	f, failures := openRecord(file, data)
	if failures != nil {
		return ledger.Goal{}, failures
	}

	state := ledger.GoalState{
		ID:        one(f, "id", ledger.ParseGoalID),
		Created:   one(f, "created", kernel.ParseCalendarDate),
		Author:    one(f, "author", kernel.ParseActor),
		Statement: one(f, "statement", kernel.ParseRecordText),
		Success:   one(f, "success", kernel.ParseRecordText),
		Nongoals:  many(f, "nongoals", kernel.ParseRecordText),
		Status:    one(f, "status", ledger.ParseGoalStatus),
		Approved:  approval(f),
	}

	return build(f, state, ledger.ValidateGoalState, ledger.RestoreGoal)
}

func decodeCapability(
	file string,
	data []byte,
) (ledger.Capability, []error) {
	f, failures := openRecord(file, data)
	if failures != nil {
		return ledger.Capability{}, failures
	}

	state := ledger.CapabilityState{
		ID:       one(f, "id", ledger.ParseCapabilityID),
		Created:  one(f, "created", kernel.ParseCalendarDate),
		Author:   one(f, "author", kernel.ParseActor),
		Name:     one(f, "name", kernel.ParseRecordText),
		Intent:   one(f, "intent", kernel.ParseRecordText),
		Paths:    many(f, "paths", kernel.ParsePathGlob),
		Serves:   one(f, "serves", ledger.ParseGoalID),
		Nongoals: many(f, "nongoals", kernel.ParseRecordText),
		Promises: promiseStates(f),
		Approved: approval(f),
	}

	return build(
		f,
		state,
		ledger.ValidateCapabilityState,
		ledger.RestoreCapability,
	)
}

func decodeDecision(file string, data []byte) (memory.Decision, []error) {
	f, failures := openRecord(file, data)
	if failures != nil {
		return memory.Decision{}, failures
	}

	state := memory.DecisionState{
		ID:         one(f, "id", memory.ParseDecisionID),
		Created:    one(f, "created", kernel.ParseCalendarDate),
		Author:     one(f, "author", kernel.ParseActor),
		What:       one(f, "what", kernel.ParseRecordText),
		Why:        one(f, "why", kernel.ParseRecordText),
		Where:      many(f, "where", kernel.ParsePathGlob),
		Supersedes: one(f, "supersedes", memory.ParseDecisionID),
		Rejected:   many(f, "rejected", kernel.ParseRecordText),
		Status:     one(f, "status", memory.ParseDecisionStatus),
		Review:     one(f, "review", memory.ParseReviewSpan),
		Reviewed:   one(f, "reviewed", kernel.ParseCalendarDate),
		Approved:   approval(f),
	}

	return build(
		f,
		state,
		memory.ValidateDecisionState,
		memory.RestoreDecision,
	)
}

func decodeQuestion(file string, data []byte) (memory.Question, []error) {
	f, failures := openRecord(file, data)
	if failures != nil {
		return memory.Question{}, failures
	}

	state := memory.QuestionState{
		ID:       one(f, "id", memory.ParseQuestionID),
		Created:  one(f, "created", kernel.ParseCalendarDate),
		Author:   one(f, "author", kernel.ParseActor),
		Text:     one(f, "text", kernel.ParseRecordText),
		Blocks:   many(f, "blocks", memory.ParseQuestionBlockID),
		Status:   one(f, "status", memory.ParseQuestionStatus),
		Answer:   one(f, "answer", kernel.ParseRecordText),
		Origin:   one(f, "origin", memory.ParseQuestionOriginID),
		Approved: approval(f),
	}

	return build(
		f,
		state,
		memory.ValidateQuestionState,
		memory.RestoreQuestion,
	)
}

func decodeGuide(file string, data []byte) (memory.Guide, []error) {
	parsed, err := ParseGuide(file, data)
	if err != nil {
		return memory.Guide{}, parseFailure(file, err)
	}

	f := newFields(file, parsed.Root)
	state := memory.GuideState{
		ID:       guideFileID(f),
		Title:    one(f, "title", kernel.ParseRecordText),
		Kind:     one(f, "kind", memory.ParseGuideKind),
		Applies:  many(f, "applies", kernel.ParsePathGlob),
		Always:   one(f, "always", parseFlag),
		Review:   one(f, "review", memory.ParseReviewSpan),
		Reviewed: one(f, "reviewed", kernel.ParseCalendarDate),
		Tags:     many(f, "tags", memory.ParseGuideTag),
		Audience: one(f, "audience", memory.ParseGuideAudience),
		Status:   one(f, "status", memory.ParseGuideStatus),
		Refers:   many(f, "refers", memory.ParseGuideCapabilityID),
		Lines:    countLines(parsed.Body),
	}

	return build(f, state, memory.ValidateGuideState, memory.RestoreGuide)
}

func countLines(body string) int {
	lines := strings.Count(body, "\n")

	if body != "" && !strings.HasSuffix(body, "\n") {
		lines++
	}

	return lines
}

func guideFileID(f *fields) memory.GuideID {
	match := guideFilePattern.FindStringSubmatch(filepath.Base(f.file))
	if match == nil {
		f.fail(1, failAt(RecordFileMismatch, f.file, 1))
		return memory.GuideID{}
	}

	id, err := memory.ParseGuideID(match[1])
	if err != nil {
		f.fail(1, err)
	}

	return id
}

func decodeWork(file string, data []byte) (agenda.Work, []error) {
	f, failures := openRecord(file, data)
	if failures != nil {
		return agenda.Work{}, failures
	}

	state := agenda.WorkState{
		ID:         one(f, "id", agenda.ParseWorkID),
		Created:    one(f, "created", kernel.ParseCalendarDate),
		Author:     one(f, "author", kernel.ParseActor),
		Request:    one(f, "request", agenda.ParseWorkRequest),
		Implements: many(f, "implements", agenda.ParseWorkPromiseID),
		Serves:     one(f, "serves", agenda.ParseWorkGoalID),
		Log:        many(f, "log", kernel.ParseRecordText),
		Status:     one(f, "status", agenda.ParseWorkStatus),
		Approved:   approval(f),
	}

	return build(f, state, agenda.ValidateWorkState, agenda.RestoreWork)
}

func decodeTask(file string, data []byte) (agenda.Task, []error) {
	f, failures := openRecord(file, data)
	if failures != nil {
		return agenda.Task{}, failures
	}

	state := agenda.TaskState{
		ID:         one(f, "id", agenda.ParseTaskID),
		Created:    one(f, "created", kernel.ParseCalendarDate),
		Author:     one(f, "author", kernel.ParseActor),
		Title:      one(f, "title", kernel.ParseRecordText),
		Parent:     one(f, "parent", agenda.ParseTaskWorkID),
		Blocks:     many(f, "blocks", agenda.ParseTaskID),
		Implements: one(f, "implements", agenda.ParseTaskPromiseID),
		Owner:      one(f, "owner", kernel.ParseActor),
		Branch:     one(f, "branch", agenda.ParseTaskBranch),
		Status:     one(f, "status", agenda.ParseTaskStatus),
		Approved:   approval(f),
	}

	return build(f, state, agenda.ValidateTaskState, agenda.RestoreTask)
}
