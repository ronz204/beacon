package record

import (
	"slices"

	yaml "go.yaml.in/yaml/v4"
)

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

func cloneNode(node *yaml.Node) *yaml.Node {
	copied := *node
	copied.Content = make([]*yaml.Node, len(node.Content))

	for index, child := range node.Content {
		copied.Content[index] = cloneNode(child)
	}

	return &copied
}
