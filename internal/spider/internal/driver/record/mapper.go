package record

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
	yaml "go.yaml.in/yaml/v4"
)

var guideFilePattern = regexp.MustCompile(
	`^(bc-[0-9a-f]{4,})-.+\.md$`,
)

type fields struct {
	file     string
	root     *yaml.Node
	values   map[string]*yaml.Node
	failures *[]error
}

func newFields(file string, root *yaml.Node) *fields {
	return &fields{
		file:     file,
		root:     root,
		values:   mappingValues(root),
		failures: &[]error{},
	}
}

func mappingValues(root *yaml.Node) map[string]*yaml.Node {
	values := make(map[string]*yaml.Node, len(root.Content)/2)

	for index := 0; index < len(root.Content); index += 2 {
		values[root.Content[index].Value] = root.Content[index+1]
	}

	return values
}

func (f *fields) fail(line int, cause error) {
	failure := kernel.NewRecordFailure(f.file, line, cause)
	*f.failures = append(*f.failures, failure)
}

func (f *fields) failType(node *yaml.Node) {
	f.fail(node.Line, failAt(RecordFieldType, f.file, node.Line))
}

func (f *fields) nested(node *yaml.Node) (*fields, bool) {
	if node.Kind != yaml.MappingNode {
		f.failType(node)
		return nil, false
	}

	return &fields{
		file:     f.file,
		root:     node,
		values:   mappingValues(node),
		failures: f.failures,
	}, true
}

func (f *fields) mapping(key string) (*fields, bool) {
	node, found := f.values[key]
	if !found {
		return nil, false
	}

	return f.nested(node)
}

func (f *fields) checkFileName() {
	node, found := f.values["id"]
	if !found || node.Kind != yaml.ScalarNode {
		return
	}

	base := filepath.Base(f.file)
	if !strings.HasPrefix(base, node.Value+"-") {
		f.fail(node.Line, failAt(RecordFileMismatch, f.file, node.Line))
	}
}

func parseScalar[T any](
	f *fields,
	node *yaml.Node,
	parse func(string) (T, error),
) T {
	var zero T

	if node.Kind != yaml.ScalarNode {
		f.failType(node)
		return zero
	}

	value, err := parse(node.Value)
	if err != nil {
		f.fail(node.Line, err)
		return zero
	}

	return value
}

func one[T any](f *fields, key string, parse func(string) (T, error)) T {
	node, found := f.values[key]
	if !found {
		var zero T
		return zero
	}

	return parseScalar(f, node, parse)
}

func many[T any](f *fields, key string, parse func(string) (T, error)) []T {
	node, found := f.values[key]
	if !found {
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		f.failType(node)
		return nil
	}

	values := make([]T, 0, len(node.Content))
	for _, item := range node.Content {
		values = append(values, parseScalar(f, item, parse))
	}

	return values
}

func rawText(raw string) (string, error) {
	return raw, nil
}

func parseFlag(raw string) (bool, error) {
	switch raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}

	return false, recordErrors.Fail(RecordFieldType, nil)
}

func build[S, T any](
	f *fields,
	state S,
	validate func(S) error,
	restore func(S) T,
) (T, []error) {
	var zero T

	if len(*f.failures) > 0 {
		return zero, *f.failures
	}
	if err := validate(state); err != nil {
		f.fail(f.root.Line, err)
		return zero, *f.failures
	}

	return restore(state), nil
}

func openRecord(file string, data []byte) (*fields, []error) {
	root, err := ParseRecord(file, data)
	if err != nil {
		return nil, parseFailure(file, err)
	}

	opened := newFields(file, root)
	opened.checkFileName()

	return opened, nil
}

func parseFailure(file string, err error) []error {
	line := 0

	var located RecordError
	if errors.As(err, &located) {
		line, _ = located.Details["line"].(int)
	}

	return []error{kernel.NewRecordFailure(file, line, err)}
}

func approval(f *fields) kernel.Approval {
	section, found := f.mapping("approved")
	if !found {
		return kernel.Approval{}
	}

	approved, err := kernel.ParseApproval(kernel.ApprovalInput{
		By:  one(section, "by", rawText),
		At:  one(section, "at", rawText),
		Via: one(section, "via", rawText),
	})
	if err != nil {
		f.fail(section.root.Line, err)
		return kernel.Approval{}
	}

	return approved
}

func attestation(f *fields) ledger.PromiseAttestation {
	section, found := f.mapping("attested")
	if !found {
		return ledger.PromiseAttestation{}
	}

	attested, err := ledger.ParsePromiseAttestation(
		ledger.PromiseAttestationInput{
			By:       one(section, "by", rawText),
			At:       one(section, "at", rawText),
			Commit:   one(section, "commit", rawText),
			Evidence: one(section, "evidence", rawText),
		},
	)
	if err != nil {
		f.fail(section.root.Line, err)
		return ledger.PromiseAttestation{}
	}

	return attested
}

func promiseCheck(f *fields) ledger.PromiseCheck {
	section, found := f.mapping("check")
	if !found {
		return ledger.PromiseCheck{}
	}

	input := ledger.PromiseCheckInput{Params: map[string]string{}}
	content := section.root.Content
	for index := 0; index < len(content); index += 2 {
		key := content[index].Value
		text := parseScalar(section, content[index+1], rawText)

		switch key {
		case "runner":
			input.Runner = text
		case "manual":
			input.Manual = text
		default:
			input.Params[key] = text
		}
	}

	check, err := ledger.ParsePromiseCheck(input)
	if err != nil {
		f.fail(section.root.Line, err)
		return ledger.PromiseCheck{}
	}

	return check
}

func promiseStates(f *fields) []ledger.PromiseState {
	node, found := f.values["promises"]
	if !found {
		line := f.root.Line
		f.fail(line, failAt(RecordFieldMissing, f.file, line))
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		f.failType(node)
		return nil
	}

	states := make([]ledger.PromiseState, 0, len(node.Content))
	for _, item := range node.Content {
		if section, isMapping := f.nested(item); isMapping {
			states = append(states, promiseState(section))
		}
	}

	return states
}

func promiseState(f *fields) ledger.PromiseState {
	return ledger.PromiseState{
		ID:       one(f, "id", ledger.ParsePromiseID),
		Text:     one(f, "text", kernel.ParseRecordText),
		Check:    promiseCheck(f),
		Status:   one(f, "status", ledger.ParsePromiseStatus),
		Approved: approval(f),
		Attested: attestation(f),
		Origin:   one(f, "origin", ledger.ParseOriginWorkID),
		Refers:   many(f, "refers", ledger.ParsePromiseID),
	}
}

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
	}

	return build(f, state, memory.ValidateGuideState, memory.RestoreGuide)
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
