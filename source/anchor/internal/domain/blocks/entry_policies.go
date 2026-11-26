package blocks

import (
	"reflect"
	"slices"
)

func MergeEntries(text string, entries []Entry) (string, error) {
	doc, err := parseDocument(text)
	if err != nil {
		return "", err
	}

	changed := false
	for _, entry := range entries {
		added, err := doc.merge(entry)
		if err != nil {
			return "", err
		}

		changed = changed || added
	}

	if !changed {
		return text, nil
	}

	return doc.String(), nil
}

func RemoveEntries(text string, entries []Entry) (string, []Entry, error) {
	doc, err := parseDocument(text)
	if err != nil {
		return "", nil, err
	}

	var kept []Entry
	changed := false

	for _, entry := range entries {
		removed, edited := doc.remove(entry)
		changed = changed || removed

		if edited {
			kept = append(kept, entry)
		}
	}

	if !changed {
		return text, kept, nil
	}

	return doc.String(), kept, nil
}

func IsEmptyDocument(text string) (bool, error) {
	doc, err := parseDocument(text)
	if err != nil {
		return false, err
	}

	return len(doc.root.keys) == 0, nil
}

func (d document) merge(entry Entry) (bool, error) {
	last := len(entry.path) - 1
	parent := d.root

	for _, key := range entry.path[:last] {
		child, exists := parent.fields[key]
		if !exists {
			child = newObject()
			parent.put(key, child)
		}
		if child.kind != objectNode {
			return false, conflictingEntry(entry.path)
		}

		parent = child
	}

	existing, exists := parent.fields[entry.path[last]]
	if entry.mode == EntrySet {
		return mergeValue(parent, entry, existing, exists)
	}

	return mergeItem(parent, entry, existing, exists)
}

func mergeValue(
	parent *node,
	entry Entry,
	existing *node,
	exists bool,
) (bool, error) {
	key := entry.path[len(entry.path)-1]

	if !exists {
		parent.put(key, nodeOf(entry.value))
		return true, nil
	}
	if reflect.DeepEqual(existing.value(), entry.value) {
		return false, nil
	}

	return false, conflictingEntry(entry.path)
}

func mergeItem(
	parent *node,
	entry Entry,
	existing *node,
	exists bool,
) (bool, error) {
	key := entry.path[len(entry.path)-1]

	if !exists {
		existing = &node{kind: arrayNode}
		parent.put(key, existing)
	}
	if existing.kind != arrayNode {
		return false, conflictingEntry(entry.path)
	}
	if containsItem(existing, entry.value) {
		return false, nil
	}

	existing.items = append(existing.items, nodeOf(entry.value))

	return true, nil
}

func containsItem(array *node, value any) bool {
	return slices.ContainsFunc(array.items, func(item *node) bool {
		return reflect.DeepEqual(item.value(), value)
	})
}

func (d document) remove(entry Entry) (bool, bool) {
	last := len(entry.path) - 1
	chain := []*node{d.root}

	for _, key := range entry.path[:last] {
		child := chain[len(chain)-1].fields[key]
		if child == nil || child.kind != objectNode {
			return false, false
		}

		chain = append(chain, child)
	}

	parent := chain[len(chain)-1]
	target, exists := parent.fields[entry.path[last]]
	if !exists {
		return false, false
	}

	removed, edited := takeOut(parent, target, entry)
	if removed {
		prune(chain, entry.path)
	}

	return removed, edited
}

func takeOut(parent, target *node, entry Entry) (bool, bool) {
	key := entry.path[len(entry.path)-1]

	if entry.mode == EntrySet {
		if !reflect.DeepEqual(target.value(), entry.value) {
			return false, true
		}

		parent.drop(key)

		return true, false
	}

	if target.kind != arrayNode {
		return false, false
	}

	before := len(target.items)
	target.items = slices.DeleteFunc(target.items, func(item *node) bool {
		return reflect.DeepEqual(item.value(), entry.value)
	})

	removed := len(target.items) < before
	if removed && len(target.items) == 0 {
		parent.drop(key)
	}

	return removed, false
}

func prune(chain []*node, path []string) {
	for position := len(chain) - 1; position > 0; position-- {
		if len(chain[position].keys) > 0 {
			return
		}

		chain[position-1].drop(path[position-1])
	}
}
