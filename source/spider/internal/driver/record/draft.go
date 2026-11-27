package record

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/system"
	yaml "go.yaml.in/yaml/v4"
)

type DraftInput struct {
	Payload []byte
	Author  kernel.Actor
	Today   kernel.CalendarDate
	Issuer  kernel.IDIssuer
	Taken   func(kernel.RecordID) bool
}

type Draft struct {
	Kind    string
	ID      kernel.RecordID
	File    string
	Before  []byte
	Output  []byte
	Entity  any
	Ignored []string
	root    *yaml.Node
	body    string
}

func (s Store) Draft(input DraftInput) (Draft, []error) {
	kind, incoming, err := parsePayload(input.Payload)
	if err != nil {
		return Draft{}, []error{err}
	}

	if kind == ConfigKind {
		return s.configDraft(incoming)
	}

	draft := Draft{Kind: kind}
	body, hasBody := takeKey(incoming, bodyKey)
	given, hasID := takeKey(incoming, idKey)
	if kind != guideKind && hasBody {
		return Draft{}, []error{payloadFailure("body")}
	}

	existing, err := s.existing(&draft, given, hasID)
	if err != nil {
		return Draft{}, []error{err}
	}

	merge := &merger{}
	draft.root = merge.mapping(existing, incoming, s.ownedKeys(kind), "")
	if kind == guideKind {
		draft.body = scalarOr(body, draft.body)
	}
	if len(merge.failures) > 0 {
		return Draft{}, merge.failures
	}
	draft.Ignored = merge.ignored

	if err := s.fill(&draft, input, merge.fresh, existing == nil); err != nil {
		return Draft{}, []error{err}
	}

	return draft.render()
}

func (s Store) ownedKeys(kind string) []string {
	if kind == guideKind {
		return nil
	}

	return recordOwned
}

func (d Draft) Stamp(stamps map[string]kernel.Approval) (Draft, []error) {
	if len(stamps) == 0 {
		return d, nil
	}

	root := cloneNode(d.root)
	target := d.ID.String()
	if d.Kind == ConfigKind {
		target = ConfigKind
	}
	targets := stampTargets(target, root)

	for _, id := range slices.Sorted(maps.Keys(stamps)) {
		target, found := targets[id]
		if !found {
			failure := recordErrors.Fail(RecordStampTarget, map[string]any{
				"id": id,
			})
			return Draft{}, []error{failure}
		}

		setApproval(target, recounted(target, stamps[id]))
	}

	d.root = root

	return d.render()
}

func stampTargets(id string, root *yaml.Node) map[string]*yaml.Node {
	targets := map[string]*yaml.Node{id: root}

	promises, found := mappingValues(root)[promisesKey]
	if found && promises.Kind == yaml.SequenceNode {
		for _, item := range promises.Content {
			targets[promiseID(item)] = item
		}
	}

	return targets
}

func scalarText(node *yaml.Node) string {
	if node == nil {
		return ""
	}

	return node.Value
}

func recounted(target *yaml.Node, approval kernel.Approval) kernel.Approval {
	section, found := mappingValues(target)[approvedKey]
	if !found || section.Kind != yaml.MappingNode {
		return approval
	}

	values := mappingValues(section)
	count, _ := strconv.Atoi(scalarText(values["n"]))
	previous, err := kernel.ParseApproval(kernel.ApprovalInput{
		By:    scalarText(values["by"]),
		At:    scalarText(values["at"]),
		Via:   scalarText(values["via"]),
		Count: count,
	})
	if err != nil || !previous.SameStamp(approval) {
		return approval
	}

	return approval.Recounted(previous.Count() + 1)
}

func setApproval(target *yaml.Node, approval kernel.Approval) {
	section := &yaml.Node{Kind: yaml.MappingNode}
	add := func(key, value string) {
		section.Content = append(
			section.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			&yaml.Node{Kind: yaml.ScalarNode, Value: value},
		)
	}

	add("by", approval.By().String())
	add("at", approval.At().String())
	if via, found := approval.Via(); found {
		add("via", via.String())
	}
	if approval.Count() > 1 {
		add("n", strconv.Itoa(approval.Count()))
	}

	takeKey(target, approvedKey)
	target.Content = append(
		target.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: approvedKey},
		section,
	)
}

func (d Draft) render() (Draft, []error) {
	raw, err := encodeDocument(&yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{cloneNode(d.root)},
	})
	if err != nil {
		return Draft{}, []error{failAt(RecordFormatUnsafe, d.File, 1)}
	}

	if d.Kind == ConfigKind {
		d.Output, err = FormatConfig(d.File, raw)
		if err != nil {
			return Draft{}, parseFailure(d.File, err)
		}
		_, failures := decodeConfig(d.File, d.Output)
		if len(failures) > 0 {
			return Draft{}, failures
		}
		return d, nil
	}

	if d.Kind == guideKind {
		raw = []byte(frontmatterOpen + string(raw) + frontmatterEnd + d.body)
		d.Output, err = FormatGuide(d.File, raw)
	} else {
		d.Output, err = FormatRecord(d.File, kindFolders[d.Kind], raw)
	}
	if err != nil {
		return Draft{}, parseFailure(d.File, err)
	}

	entity, failures := decoders[d.Kind](d.File, d.Output)
	if len(failures) > 0 {
		return Draft{}, failures
	}
	d.Entity = entity

	return d, nil
}

func (d Draft) Unchanged() bool {
	return bytes.Equal(d.Before, d.Output)
}

func (s Store) Write(d Draft, lock string) error {
	return system.WithFileLock(lock, d.Commit)
}

func (d Draft) Commit() error {
	if err := Unchanged(d.File, d.Before); err != nil {
		return err
	}

	if d.Unchanged() {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(d.File), 0o755); err != nil {
		return kernel.NewRecordFailure(d.File, 0, err)
	}
	if err := system.WriteFileAtomic(d.File, d.Output); err != nil {
		return kernel.NewRecordFailure(d.File, 0, err)
	}

	return nil
}

func (s Store) configDraft(incoming *yaml.Node) (Draft, []error) {
	draft := Draft{Kind: ConfigKind, File: s.ConfigFile()}

	data, err := os.ReadFile(draft.File)
	if err != nil {
		return Draft{}, []error{kernel.NewRecordFailure(draft.File, 0, err)}
	}
	draft.Before = data

	existing, err := ParseRecord(draft.File, data)
	if err != nil {
		return Draft{}, parseFailure(draft.File, err)
	}

	merge := &merger{}
	draft.root = merge.mapping(existing, incoming, []string{approvedKey}, "")
	if len(merge.failures) > 0 {
		return Draft{}, merge.failures
	}
	draft.Ignored = merge.ignored

	if sameNode(existing, draft.root) {
		draft.Output = data
		return draft, nil
	}

	return draft.render()
}
