package record

import (
	"os"

	"github.com/ronz204/beacon/source/kernel"
	yaml "go.yaml.in/yaml/v4"
)

func (s Store) existing(
	draft *Draft,
	given *yaml.Node,
	hasID bool,
) (*yaml.Node, error) {
	if !hasID {
		return nil, nil
	}
	if given.Kind != yaml.ScalarNode {
		return nil, payloadFailure(idKey)
	}

	id, err := kernel.ParseRecordID(given.Value)
	if err != nil {
		return nil, payloadFailure(idKey).WithCause(err)
	}
	draft.ID = id

	file, err := s.findFile(id)
	if err != nil {
		if draft.Kind == "task" {
			return nil, nil
		}
		return nil, err
	}

	return s.readExisting(draft, file)
}

func (s Store) readExisting(draft *Draft, file string) (*yaml.Node, error) {
	kind, _ := KindOfFile(file)
	if kind != draft.Kind {
		return nil, payloadFailure(payloadTypeKey)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, kernel.NewRecordFailure(file, 0, err)
	}
	draft.File = file
	draft.Before = data

	if kind != guideKind {
		return ParseRecord(file, data)
	}

	parsed, err := ParseGuide(file, data)
	if err != nil {
		return nil, err
	}
	draft.body = parsed.Body

	return parsed.Root, nil
}
