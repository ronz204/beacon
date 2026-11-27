package record

import (
	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	yaml "go.yaml.in/yaml/v4"
)

const attestedKey = "attested"

func (s Store) DraftAttesting(
	file string,
	promise string,
	attestation ledger.PromiseAttestation,
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

	target, found := promiseNode(root, promise)
	if !found {
		return Draft{}, []error{MissingFailure(promise)}
	}

	setAttestation(target, attestation)
	draft.root = root

	return draft.render()
}

func setAttestation(target *yaml.Node, attestation ledger.PromiseAttestation) {
	section := &yaml.Node{Kind: yaml.MappingNode}
	add := func(key, value string) {
		section.Content = append(
			section.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			&yaml.Node{Kind: yaml.ScalarNode, Value: value},
		)
	}

	add("by", attestation.By().String())
	add("at", attestation.At().String())
	if via, found := attestation.Via(); found {
		add("via", via.String())
	}
	add("commit", attestation.Commit())
	if attestation.Evidence() != "" {
		add("evidence", attestation.Evidence())
	}

	takeKey(target, attestedKey)
	target.Content = append(
		target.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: attestedKey},
		section,
	)
}
