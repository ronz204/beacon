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

func (r ChangedRecord) promiseByID(id kernel.RecordID) (ChangedPromise, bool) {
	for _, promise := range r.promises {
		if promise.id == id {
			return promise, true
		}
	}

	return ChangedPromise{}, false
}

type StateChangeInput struct {
	Path     string
	Kind     string
	Baseline bool
	Runners  []string
	Before   *ChangedRecordInput
	After    *ChangedRecordInput
}

type StateChange struct {
	path     string
	kind     ChangeKind
	baseline bool
	runners  []string
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

	if !input.Baseline && before == nil && after == nil {
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
		before:   before,
		after:    after,
	}, nil
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

func (c StateChange) Runners() []string {
	return slices.Clone(c.runners)
}
