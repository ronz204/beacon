package blocks

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"sort"
	"strings"
)

type EntryMode string

const (
	EntrySet    EntryMode = "set"
	EntryAppend EntryMode = "append"
)

type Entry struct {
	mode  EntryMode
	path  []string
	value any
}

func ParseEntry(mode EntryMode, value any, path ...string) (Entry, error) {
	if mode != EntrySet && mode != EntryAppend {
		return Entry{}, invalidEntry(path)
	}
	if len(path) == 0 || slices.Contains(path, "") {
		return Entry{}, invalidEntry(path)
	}

	canonical, err := canonicalValue(value)
	if err != nil {
		return Entry{}, invalidEntry(path).WithCause(err)
	}

	return Entry{mode: mode, path: slices.Clone(path), value: canonical}, nil
}

func (e Entry) Mode() EntryMode {
	return e.mode
}

func (e Entry) Path() []string {
	return slices.Clone(e.path)
}

func canonicalValue(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var canonical any
	err = json.Unmarshal(data, &canonical)

	return canonical, err
}

type nodeKind int

const (
	scalarNode nodeKind = iota
	objectNode
	arrayNode
)

type node struct {
	kind   nodeKind
	raw    string
	keys   []string
	fields map[string]*node
	items  []*node
}

func newObject() *node {
	return &node{kind: objectNode, fields: map[string]*node{}}
}

func (n *node) put(key string, child *node) {
	if _, exists := n.fields[key]; !exists {
		n.keys = append(n.keys, key)
	}

	n.fields[key] = child
}

func (n *node) drop(key string) {
	delete(n.fields, key)
	n.keys = slices.DeleteFunc(n.keys, func(k string) bool {
		return k == key
	})
}

func (n *node) value() any {
	switch n.kind {
	case objectNode:
		result := make(map[string]any, len(n.keys))
		for key, child := range n.fields {
			result[key] = child.value()
		}
		return result
	case arrayNode:
		result := make([]any, 0, len(n.items))
		for _, child := range n.items {
			result = append(result, child.value())
		}
		return result
	default:
		var result any
		_ = json.Unmarshal([]byte(n.raw), &result)
		return result
	}
}

func nodeOf(value any) *node {
	switch typed := value.(type) {
	case map[string]any:
		object := newObject()
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			object.put(key, nodeOf(typed[key]))
		}
		return object
	case []any:
		array := &node{kind: arrayNode}
		for _, item := range typed {
			array.items = append(array.items, nodeOf(item))
		}
		return array
	default:
		return &node{kind: scalarNode, raw: scalarText(value)}
	}
}

func scalarText(value any) string {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)

	return strings.TrimSuffix(buffer.String(), "\n")
}

func (n *node) write(out *strings.Builder, indent string, depth int) {
	switch {
	case n.kind == objectNode && len(n.keys) > 0:
		n.writeObject(out, indent, depth)
	case n.kind == objectNode:
		out.WriteString("{}")
	case n.kind == arrayNode && len(n.items) > 0:
		n.writeArray(out, indent, depth)
	case n.kind == arrayNode:
		out.WriteString("[]")
	default:
		out.WriteString(n.raw)
	}
}

func (n *node) writeObject(out *strings.Builder, indent string, depth int) {
	out.WriteString("{\n")

	for position, key := range n.keys {
		out.WriteString(strings.Repeat(indent, depth+1))
		out.WriteString(scalarText(key))
		out.WriteString(": ")
		n.fields[key].write(out, indent, depth+1)
		out.WriteString(separator(position, len(n.keys)))
	}

	out.WriteString(strings.Repeat(indent, depth) + "}")
}

func (n *node) writeArray(out *strings.Builder, indent string, depth int) {
	out.WriteString("[\n")

	for position, item := range n.items {
		out.WriteString(strings.Repeat(indent, depth+1))
		item.write(out, indent, depth+1)
		out.WriteString(separator(position, len(n.items)))
	}

	out.WriteString(strings.Repeat(indent, depth) + "]")
}

func separator(position, total int) string {
	if position < total-1 {
		return ",\n"
	}

	return "\n"
}

type document struct {
	root     *node
	indent   string
	newline  string
	trailing bool
}

func parseDocument(text string) (document, error) {
	if strings.TrimSpace(text) == "" {
		return document{
			root:     newObject(),
			indent:   "  ",
			newline:  "\n",
			trailing: true,
		}, nil
	}

	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()

	root, err := parseNode(decoder)
	if err == nil && root.kind != objectNode {
		err = io.ErrUnexpectedEOF
	}
	if err == nil {
		_, err = decoder.Token()
		err = trailingData(err)
	}
	if err != nil {
		return document{}, invalidDocument(err)
	}

	return document{
		root:     root,
		indent:   detectIndent(text),
		newline:  lineEnding(text),
		trailing: strings.HasSuffix(text, "\n"),
	}, nil
}

func trailingData(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return io.ErrUnexpectedEOF
	}

	return err
}

func parseNode(decoder *json.Decoder) (*node, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	delimiter, isDelimiter := token.(json.Delim)
	switch {
	case !isDelimiter:
		return &node{kind: scalarNode, raw: scalarText(token)}, nil
	case delimiter == '{':
		return parseObject(decoder)
	default:
		return parseArray(decoder)
	}
}

func parseObject(decoder *json.Decoder) (*node, error) {
	object := newObject()

	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		child, err := parseNode(decoder)
		if err != nil {
			return nil, err
		}

		object.put(key.(string), child)
	}

	_, err := decoder.Token()

	return object, err
}

func parseArray(decoder *json.Decoder) (*node, error) {
	array := &node{kind: arrayNode}

	for decoder.More() {
		child, err := parseNode(decoder)
		if err != nil {
			return nil, err
		}

		array.items = append(array.items, child)
	}

	_, err := decoder.Token()

	return array, err
}

func detectIndent(text string) string {
	for line := range strings.SplitSeq(text, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" || trimmed == line {
			continue
		}

		return line[:len(line)-len(trimmed)]
	}

	return "  "
}

func (d document) String() string {
	var out strings.Builder
	d.root.write(&out, d.indent, 0)

	if d.trailing {
		out.WriteString("\n")
	}

	return strings.ReplaceAll(out.String(), "\n", d.newline)
}
