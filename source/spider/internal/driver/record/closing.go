package record

import (
	"bytes"
	"errors"
	"io/fs"
	"os"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/system"
	yaml "go.yaml.in/yaml/v4"
)

const closedFile = "closed.json"

type Closing struct {
	Closed     kernel.CalendarDate
	Reason     kernel.RecordText
	Capability string
	Approval   *kernel.Approval
}

func ClosedJSON(record []byte, closing Closing) ([]byte, error) {
	root, err := ParseRecord(closedFile, record)
	if err != nil {
		return nil, err
	}

	for _, key := range []string{"closed", "reason", "capability"} {
		takeKey(root, key)
	}

	appendScalar(root, "closed", closing.Closed.String())
	appendScalar(root, "reason", closing.Reason.String())
	if closing.Capability != "" {
		appendScalar(root, "capability", closing.Capability)
	}
	if closing.Approval != nil {
		setApproval(root, *closing.Approval)
	}

	return objectJSON(root), nil
}

func appendScalar(mapping *yaml.Node, key, value string) {
	mapping.Content = append(
		mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: value,
			Style: yaml.DoubleQuotedStyle,
		},
	)
}

func PromiseJSON(file string, data []byte, promise string) ([]byte, bool) {
	root, err := ParseRecord(file, data)
	if err != nil {
		return nil, false
	}

	item, found := promiseNode(root, promise)
	if !found {
		return nil, false
	}

	return objectJSON(item), true
}

func promiseNode(root *yaml.Node, promise string) (*yaml.Node, bool) {
	promises, found := mappingValues(root)[promisesKey]
	if !found || promises.Kind != yaml.SequenceNode {
		return nil, false
	}

	for _, item := range promises.Content {
		if promiseID(item) == promise {
			return item, true
		}
	}

	return nil, false
}

func (s Store) DraftWithoutPromise(
	file string,
	promise string,
) (Draft, []error) {
	draft := Draft{Kind: "capability"}

	root, err := s.readExisting(&draft, file)
	if err != nil {
		return Draft{}, []error{err}
	}

	id, err := kernel.ParseRecordID(scalarOr(mappingValues(root)[idKey], ""))
	if err != nil {
		return Draft{}, []error{err}
	}
	draft.ID = id

	promises := mappingValues(root)[promisesKey]
	promises.Content = slicesWithout(promises.Content, promise)
	draft.root = root

	return draft.render()
}

func slicesWithout(items []*yaml.Node, promise string) []*yaml.Node {
	kept := make([]*yaml.Node, 0, len(items))

	for _, item := range items {
		if promiseID(item) != promise {
			kept = append(kept, item)
		}
	}

	return kept
}

func Unchanged(file string, before []byte) error {
	current, err := os.ReadFile(file)
	switch {
	case errors.Is(err, fs.ErrNotExist) && before == nil:
		return nil
	case err != nil:
		return kernel.NewRecordFailure(file, 0, err)
	case !bytes.Equal(current, before):
		return failAt(RecordChanged, file, 0)
	}

	return nil
}

func Remove(file string, before []byte) error {
	if err := Unchanged(file, before); err != nil {
		return err
	}

	if err := system.RemoveFile(file); err != nil {
		return kernel.NewRecordFailure(file, 0, err)
	}

	return nil
}

func StampTargetFailure(id string) error {
	return recordErrors.Fail(RecordStampTarget, map[string]any{"id": id})
}
