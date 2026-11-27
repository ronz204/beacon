package record

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	yaml "go.yaml.in/yaml/v4"
)

var errApprovalCount = errors.New(
	"approval n must be a whole number from 1",
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

	if located, ok := errors.AsType[RecordError](err); ok {
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
		By:    one(section, "by", rawText),
		At:    one(section, "at", rawText),
		Via:   one(section, "via", rawText),
		Count: one(section, "n", approvalCount),
	})
	if err != nil {
		f.fail(section.root.Line, err)
		return kernel.Approval{}
	}

	return approved
}

func approvalCount(raw string) (int, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return 0, nil
	}

	count, err := strconv.Atoi(text)
	if err != nil || count < 1 {
		return 0, errApprovalCount
	}

	return count, nil
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
			Via:      one(section, "via", rawText),
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
