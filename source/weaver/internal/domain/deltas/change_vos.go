package deltas

import (
	"maps"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeModified ChangeKind = "modified"
	ChangeDeleted  ChangeKind = "deleted"
)

func ChangeKinds() []ChangeKind {
	return []ChangeKind{ChangeAdded, ChangeModified, ChangeDeleted}
}

func ParseChangeKind(raw string) (ChangeKind, error) {
	kind := ChangeKind(strings.TrimSpace(raw))

	if !slices.Contains(ChangeKinds(), kind) {
		return "", invalidChangeValue(ChangeKindInvalid, raw)
	}

	return kind, nil
}

type ChangeLevel int

const (
	LevelMaintenance ChangeLevel = 1
	LevelAddition    ChangeLevel = 2
	LevelWeakening   ChangeLevel = 3
)

type RecordKind string

const (
	RecordGoal       RecordKind = "goal"
	RecordCapability RecordKind = "capability"
	RecordDecision   RecordKind = "decision"
	RecordQuestion   RecordKind = "question"
	RecordGuide      RecordKind = "guide"
	RecordWork       RecordKind = "work"
	RecordTask       RecordKind = "task"
	RecordConfig     RecordKind = "config"
)

func RecordKinds() []RecordKind {
	return []RecordKind{
		RecordGoal,
		RecordCapability,
		RecordDecision,
		RecordQuestion,
		RecordGuide,
		RecordWork,
		RecordTask,
		RecordConfig,
	}
}

func ParseRecordKind(raw string) (RecordKind, error) {
	kind := RecordKind(strings.TrimSpace(raw))

	if !slices.Contains(RecordKinds(), kind) {
		return "", invalidChangeValue(ChangeRecordKindInvalid, raw)
	}

	return kind, nil
}

type ChangeReason string

const (
	ReasonPromiseText      ChangeReason = "promise_text_changed"
	ReasonPromiseCheck     ChangeReason = "promise_check_changed"
	ReasonPromiseDemoted   ChangeReason = "promise_demoted"
	ReasonPromiseActivated ChangeReason = "promise_activated"
	ReasonPromiseNewActive ChangeReason = "promise_added_active"
	ReasonPromiseClosed    ChangeReason = "promise_closed"
	ReasonGoalAdded        ChangeReason = "goal_added"
	ReasonGoalChanged      ChangeReason = "goal_changed"
	ReasonGoalClosed       ChangeReason = "goal_closed"
	ReasonRunnersChanged   ChangeReason = "runners_changed"
	ReasonSchemaChanged    ChangeReason = "schema_changed"
	ReasonPacksChanged     ChangeReason = "packs_changed"
	ReasonBaselineChanged  ChangeReason = "baseline_changed"
)

type TargetKind string

const (
	TargetPromise  TargetKind = "promise"
	TargetRecord   TargetKind = "record"
	TargetConfig   TargetKind = "config"
	TargetBaseline TargetKind = "baseline"
)

type ChangeTarget struct {
	kind  TargetKind
	value string
}

func PromiseTarget(id kernel.RecordID) ChangeTarget {
	return ChangeTarget{kind: TargetPromise, value: id.String()}
}

func RecordTarget(id kernel.RecordID) ChangeTarget {
	return ChangeTarget{kind: TargetRecord, value: id.String()}
}

func ConfigTarget() ChangeTarget {
	return ChangeTarget{kind: TargetConfig}
}

func BaselineTarget(path string) (ChangeTarget, error) {
	value := strings.TrimSpace(path)

	if value == "" {
		return ChangeTarget{}, invalidChangeValue(ChangeTargetInvalid, path)
	}

	return ChangeTarget{kind: TargetBaseline, value: value}, nil
}

func (t ChangeTarget) Kind() TargetKind {
	return t.kind
}

func (t ChangeTarget) Value() string {
	return t.value
}

type ChangedPromiseInput struct {
	ID       string
	Text     string
	Check    string
	Runner   string
	Status   string
	Approved string
}

type ChangedPromise struct {
	id       kernel.RecordID
	text     string
	check    string
	runner   string
	active   bool
	approved string
}

func parseChangedPromise(input ChangedPromiseInput) (ChangedPromise, error) {
	id, err := kernel.ParseRootRecordID(input.ID)
	if err != nil {
		return ChangedPromise{}, err
	}

	if input.Status != "active" && input.Status != "proposed" {
		return ChangedPromise{}, invalidChangeValue(
			ChangeRecordInvalid,
			input.Status,
		)
	}

	return ChangedPromise{
		id:       id,
		text:     input.Text,
		check:    input.Check,
		runner:   input.Runner,
		active:   input.Status == "active",
		approved: input.Approved,
	}, nil
}

type ChangedRecordInput struct {
	Kind     string
	ID       string
	Fields   map[string]string
	Approved string
	Promises []ChangedPromiseInput
}

type ChangedRecord struct {
	kind     RecordKind
	id       kernel.RecordID
	fields   map[string]string
	approved string
	promises []ChangedPromise
}

func ParseChangedRecord(input ChangedRecordInput) (ChangedRecord, error) {
	kind, err := ParseRecordKind(input.Kind)
	if err != nil {
		return ChangedRecord{}, err
	}

	id, err := kernel.ParseRecordID(input.ID)
	if kind != RecordConfig && err != nil {
		return ChangedRecord{}, err
	}

	promises := make([]ChangedPromise, 0, len(input.Promises))
	for _, raw := range input.Promises {
		promise, err := parseChangedPromise(raw)
		if err != nil {
			return ChangedRecord{}, err
		}

		promises = append(promises, promise)
	}

	return ChangedRecord{
		kind:     kind,
		id:       id,
		fields:   maps.Clone(input.Fields),
		approved: input.Approved,
		promises: promises,
	}, nil
}

func (r ChangedRecord) Kind() RecordKind {
	return r.kind
}

func (r ChangedRecord) ID() kernel.RecordID {
	return r.id
}

func (p ChangedPromise) sameContent(other ChangedPromise) bool {
	if p.text != other.text || p.check != other.check {
		return false
	}

	return p.active == other.active
}

func (r ChangedRecord) promiseByID(id kernel.RecordID) (ChangedPromise, bool) {
	for _, promise := range r.promises {
		if promise.id == id {
			return promise, true
		}
	}

	return ChangedPromise{}, false
}

type ClosureInput struct {
	ID       string
	Approved string
	Promises []string
}

type StateChangeInput struct {
	Path     string
	Kind     string
	Baseline bool
	Runners  []string
	Closed   []ClosureInput
	Before   *ChangedRecordInput
	After    *ChangedRecordInput
}

type StateChange struct {
	path     string
	kind     ChangeKind
	baseline bool
	runners  []string
	closed   []closure
	before   *ChangedRecord
	after    *ChangedRecord
}

func ParseStateChange(input StateChangeInput) (StateChange, error) {
	kind, err := ParseChangeKind(input.Kind)
	if err != nil {
		return StateChange{}, err
	}

	before, err := parseOptionalRecord(input.Before)
	if err != nil {
		return StateChange{}, err
	}

	after, err := parseOptionalRecord(input.After)
	if err != nil {
		return StateChange{}, err
	}

	closed, err := parseClosures(input.Closed)
	if err != nil {
		return StateChange{}, err
	}

	if !isStandalone(input, closed) && before == nil && after == nil {
		return StateChange{}, invalidChangeValue(
			ChangeRecordInvalid,
			input.Path,
		)
	}

	return StateChange{
		path:     input.Path,
		kind:     kind,
		baseline: input.Baseline,
		runners:  slices.Clone(input.Runners),
		closed:   closed,
		before:   before,
		after:    after,
	}, nil
}

type closure struct {
	id       kernel.RecordID
	approved string
	promises []kernel.RecordID
}

func parseClosures(inputs []ClosureInput) ([]closure, error) {
	closures := make([]closure, 0, len(inputs))

	for _, input := range inputs {
		id, err := kernel.ParseRecordID(input.ID)
		if err != nil {
			return nil, err
		}

		promises := make([]kernel.RecordID, 0, len(input.Promises))
		for _, raw := range input.Promises {
			promise, err := kernel.ParseRecordID(raw)
			if err != nil {
				return nil, err
			}

			promises = append(promises, promise)
		}

		closures = append(closures, closure{
			id:       id,
			approved: input.Approved,
			promises: promises,
		})
	}

	return closures, nil
}

func isStandalone(input StateChangeInput, closed []closure) bool {
	return input.Baseline || len(closed) > 0
}

func parseOptionalRecord(input *ChangedRecordInput) (*ChangedRecord, error) {
	if input == nil {
		return nil, nil
	}

	parsed, err := ParseChangedRecord(*input)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func (c StateChange) Path() string {
	return c.path
}

func (c StateChange) Kind() ChangeKind {
	return c.kind
}

func (c StateChange) Baseline() bool {
	return c.baseline
}

func (c StateChange) PromisesUsing(runners []string) []string {
	if c.after == nil {
		return nil
	}

	var ids []string
	for _, promise := range c.after.promises {
		if promise.runner != "" && slices.Contains(runners, promise.runner) {
			ids = append(ids, promise.id.String())
		}
	}

	return ids
}

func (c StateChange) AddedPromises() []string {
	if c.after == nil {
		return nil
	}

	var known []kernel.RecordID
	if c.before != nil {
		for _, promise := range c.before.promises {
			known = append(known, promise.id)
		}
	}

	var ids []string
	for _, promise := range c.after.promises {
		if !slices.Contains(known, promise.id) {
			ids = append(ids, promise.id.String())
		}
	}

	return ids
}

func (c StateChange) Runners() []string {
	return slices.Clone(c.runners)
}
