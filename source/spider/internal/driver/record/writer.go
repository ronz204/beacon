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

const (
	payloadFile     = "payload.json"
	promisesKey     = "promises"
	approvedKey     = "approved"
	payloadTypeKey  = "type"
	payloadDataKey  = "record"
	payloadFileKey  = "file"
	payloadArchived = "archived"
)

var (
	recordOwned  = []string{idKey, "created", "author", approvedKey}
	promiseOwned = []string{idKey, approvedKey, "attested"}
	payloadKeys  = []string{
		payloadTypeKey,
		payloadDataKey,
		payloadFileKey,
		payloadArchived,
	}
)

var titleKeys = map[string]string{
	"goal":       "statement",
	"capability": "name",
	"decision":   "what",
	"question":   "text",
	guideKind:    "title",
	"work":       "request",
	"task":       "title",
}

var decoders = map[string]func(string, []byte) (any, []error){
	"goal":       entityOf(decodeGoal),
	"capability": entityOf(decodeCapability),
	"decision":   entityOf(decodeDecision),
	"question":   entityOf(decodeQuestion),
	guideKind:    entityOf(decodeGuide),
	"work":       entityOf(decodeWork),
	"task":       entityOf(decodeTask),
}

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

type merger struct {
	ignored  []string
	failures []error
	fresh    []*yaml.Node
}

func entityOf[T any](decode decoder[T]) func(string, []byte) (any, []error) {
	return func(file string, data []byte) (any, []error) {
		value, failures := decode(file, data)
		return value, failures
	}
}

func (s Store) Draft(input DraftInput) (Draft, []error) {
	kind, incoming, err := parsePayload(input.Payload)
	if err != nil {
		return Draft{}, []error{err}
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

func parsePayload(data []byte) (string, *yaml.Node, error) {
	root, err := ParseRecord(payloadFile, data)
	if err != nil {
		return "", nil, payloadFailure("").WithCause(err)
	}

	values := mappingValues(root)
	for key := range values {
		if !slices.Contains(payloadKeys, key) {
			return "", nil, payloadFailure(key)
		}
	}

	kindNode, hasKind := values[payloadTypeKey]
	if !hasKind || kindNode.Kind != yaml.ScalarNode {
		return "", nil, payloadFailure(payloadTypeKey)
	}
	if _, known := kindFolders[kindNode.Value]; !known {
		return "", nil, payloadFailure(payloadTypeKey)
	}

	recordNode, hasRecord := values[payloadDataKey]
	if !hasRecord || recordNode.Kind != yaml.MappingNode {
		return "", nil, payloadFailure(payloadDataKey)
	}

	return kindNode.Value, recordNode, nil
}

func payloadFailure(field string) RecordError {
	return recordErrors.Fail(RecordPayloadInvalid, map[string]any{
		"field": field,
	})
}

func takeKey(mapping *yaml.Node, key string) (*yaml.Node, bool) {
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			value := mapping.Content[index+1]
			mapping.Content = slices.Delete(mapping.Content, index, index+2)
			return value, true
		}
	}

	return nil, false
}

func scalarOr(node *yaml.Node, fallback string) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return fallback
	}

	return node.Value
}

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

func (m *merger) mapping(
	existing, incoming *yaml.Node,
	owned []string,
	path string,
) *yaml.Node {
	result := &yaml.Node{Kind: yaml.MappingNode}
	if existing != nil {
		result.Style = existing.Style
		result.HeadComment = existing.HeadComment
		result.FootComment = existing.FootComment
	}

	given := mappingValues(incoming)
	seen := make(map[string]bool)

	var pairs []*yaml.Node
	if existing != nil {
		pairs = existing.Content
	}

	for index := 0; index < len(pairs); index += 2 {
		key, value := pairs[index], pairs[index+1]
		seen[key.Value] = true
		next, found := given[key.Value]

		switch {
		case slices.Contains(owned, key.Value):
			if found && !sameNode(value, next) {
				m.ignored = append(m.ignored, path+key.Value)
			}
			result.Content = append(result.Content, key, value)
		case found:
			merged := m.value(value, next, path+key.Value)
			result.Content = append(result.Content, key, merged)
		}
	}

	for index := 0; index < len(incoming.Content); index += 2 {
		key, value := incoming.Content[index], incoming.Content[index+1]

		switch {
		case seen[key.Value]:
			continue
		case slices.Contains(owned, key.Value):
			m.ignored = append(m.ignored, path+key.Value)
		case key.Value == promisesKey:
			merged := m.value(nil, value, path+key.Value)
			result.Content = append(result.Content, key, merged)
		default:
			result.Content = append(result.Content, key, value)
		}
	}

	return result
}

func (m *merger) value(existing, incoming *yaml.Node, path string) *yaml.Node {
	switch {
	case path == promisesKey:
		return m.promises(existing, incoming)
	case existing == nil:
		return incoming
	case sameNode(existing, incoming):
		return existing
	case bothKind(existing, incoming, yaml.MappingNode):
		return m.mapping(existing, incoming, nil, path+".")
	}

	return incoming
}

func bothKind(a, b *yaml.Node, kind yaml.Kind) bool {
	return a.Kind == kind && b.Kind == kind
}

func (m *merger) promises(existing, incoming *yaml.Node) *yaml.Node {
	if incoming.Kind != yaml.SequenceNode {
		m.failures = append(m.failures, payloadFailure(promisesKey))
		return incoming
	}

	current := make(map[string]*yaml.Node)
	if existing != nil && existing.Kind == yaml.SequenceNode {
		for _, item := range existing.Content {
			current[promiseID(item)] = item
		}
	}

	result := &yaml.Node{Kind: yaml.SequenceNode}
	kept := make(map[string]bool)

	for _, item := range incoming.Content {
		if item.Kind != yaml.MappingNode {
			m.failures = append(m.failures, payloadFailure(promisesKey))
			continue
		}

		id := promiseID(item)
		previous, found := current[id]
		switch {
		case id == "":
			merged := m.mapping(nil, item, promiseOwned, "promises.")
			m.fresh = append(m.fresh, merged)
			result.Content = append(result.Content, merged)
		case found:
			kept[id] = true
			path := "promises." + id + "."
			merged := m.mapping(previous, item, promiseOwned, path)
			result.Content = append(result.Content, merged)
		default:
			m.failures = append(m.failures, promiseFailure(
				RecordPromiseUnknown,
				[]string{id},
			))
		}
	}

	var omitted []string
	for id := range current {
		if !kept[id] {
			omitted = append(omitted, id)
		}
	}
	if len(omitted) > 0 {
		slices.Sort(omitted)
		m.failures = append(
			m.failures,
			promiseFailure(RecordPromiseOmitted, omitted),
		)
	}

	return result
}

func promiseFailure(code RecordCode, ids []string) RecordError {
	return recordErrors.Fail(code, map[string]any{"promises": ids})
}

func sameNode(a, b *yaml.Node) bool {
	if a.Kind != b.Kind || len(a.Content) != len(b.Content) {
		return false
	}

	switch a.Kind {
	case yaml.ScalarNode:
		return a.Value == b.Value
	case yaml.SequenceNode:
		for index := range a.Content {
			if !sameNode(a.Content[index], b.Content[index]) {
				return false
			}
		}
		return true
	case yaml.MappingNode:
		return sameMapping(a, b)
	}

	return false
}

func sameMapping(a, b *yaml.Node) bool {
	others := mappingValues(b)

	for index := 0; index < len(a.Content); index += 2 {
		other, found := others[a.Content[index].Value]
		if !found || !sameNode(a.Content[index+1], other) {
			return false
		}
	}

	return true
}

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

func (d Draft) Stamp(stamps map[string]kernel.Approval) (Draft, []error) {
	if len(stamps) == 0 {
		return d, nil
	}

	root := cloneNode(d.root)
	targets := stampTargets(d.ID.String(), root)

	for _, id := range slices.Sorted(maps.Keys(stamps)) {
		target, found := targets[id]
		if !found {
			failure := recordErrors.Fail(RecordStampTarget, map[string]any{
				"id": id,
			})
			return Draft{}, []error{failure}
		}

		setApproval(target, stamps[id])
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

	takeKey(target, approvedKey)
	target.Content = append(
		target.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: approvedKey},
		section,
	)
}

func cloneNode(node *yaml.Node) *yaml.Node {
	copied := *node
	copied.Content = make([]*yaml.Node, len(node.Content))

	for index, child := range node.Content {
		copied.Content[index] = cloneNode(child)
	}

	return &copied
}

func (d Draft) render() (Draft, []error) {
	raw, err := encodeDocument(&yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{cloneNode(d.root)},
	})
	if err != nil {
		return Draft{}, []error{failAt(RecordFormatUnsafe, d.File, 1)}
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
