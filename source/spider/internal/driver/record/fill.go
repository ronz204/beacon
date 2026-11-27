package record

import (
	"path/filepath"
	"slices"
	"strconv"

	"github.com/ronz204/beacon/source/kernel"
	yaml "go.yaml.in/yaml/v4"
)

func (s Store) fill(
	draft *Draft,
	input DraftInput,
	fresh []*yaml.Node,
	created bool,
) error {
	taken := takenBy(input.Taken)

	if created {
		if err := s.assignID(draft, input, taken); err != nil {
			return err
		}
		if draft.Kind != guideKind {
			prepend(draft.root, "author", input.Author.String())
			prepend(draft.root, "created", input.Today.String())
			prepend(draft.root, idKey, draft.ID.String())
		}
		draft.File = s.newFile(draft)
	}

	for _, promise := range fresh {
		id, err := input.Issuer.IssueRecordID(taken.check)
		if err != nil {
			return err
		}
		taken.add(id)
		prepend(promise, idKey, id.String())
	}

	return nil
}

type takenIDs struct {
	external func(kernel.RecordID) bool
	local    map[string]bool
}

func takenBy(external func(kernel.RecordID) bool) *takenIDs {
	return &takenIDs{external: external, local: make(map[string]bool)}
}

func (t *takenIDs) check(id kernel.RecordID) bool {
	return t.local[id.String()] || t.external(id)
}

func (t *takenIDs) add(id kernel.RecordID) {
	t.local[id.String()] = true
}

func (s Store) assignID(draft *Draft, input DraftInput, taken *takenIDs) error {
	if draft.Kind == "task" {
		return assignTaskID(draft, taken)
	}

	id, err := input.Issuer.IssueRecordID(taken.check)
	if err != nil {
		return err
	}

	draft.ID = id
	taken.add(id)

	return nil
}

func assignTaskID(draft *Draft, taken *takenIDs) error {
	if draft.ID.String() != "" {
		if taken.check(draft.ID) {
			return payloadFailure(idKey)
		}
		return nil
	}

	parent, found := mappingValues(draft.root)["parent"]
	if !found || parent.Kind != yaml.ScalarNode {
		return payloadFailure("parent")
	}

	root, err := kernel.ParseRootRecordID(parent.Value)
	if err != nil {
		return payloadFailure("parent").WithCause(err)
	}

	for child := 1; ; child++ {
		raw := root.String() + "." + strconv.Itoa(child)
		id, err := kernel.ParseRecordID(raw)
		if err != nil {
			return err
		}

		if !taken.check(id) {
			draft.ID = id
			taken.add(id)
			return nil
		}
	}
}

func (s Store) newFile(draft *Draft) string {
	title := ""
	if node, found := mappingValues(draft.root)[titleKeys[draft.Kind]]; found {
		title = node.Value
	}

	name := draft.ID.String() + "-" + slugOf(title) + extensionOf(draft.Kind)

	return filepath.Join(s.folder(kindFolders[draft.Kind]), name)
}

func prepend(mapping *yaml.Node, key, value string) {
	mapping.Content = slices.Insert(
		mapping.Content,
		0,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
}
