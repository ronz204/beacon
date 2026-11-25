package record

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	yaml "go.yaml.in/yaml/v4"
)

const (
	guideKind = "guide"
	bodyKey   = "body"
	idKey     = "id"
)

var kindFolders = map[string]string{
	"goal":       "goals",
	"capability": "capabilities",
	"decision":   "decisions",
	"question":   "questions",
	guideKind:    "guides",
	"work":       "works",
	"task":       "tasks",
}

var decimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)

var quotedStyles = yaml.DoubleQuotedStyle |
	yaml.SingleQuotedStyle |
	yaml.LiteralStyle |
	yaml.FoldedStyle

type Document struct {
	Kind string
	File string
	JSON []byte
}

func RecordKinds() []string {
	return slices.Sorted(func(yield func(string) bool) {
		for kind := range kindFolders {
			if !yield(kind) {
				return
			}
		}
	})
}

func KindOfFile(file string) (string, bool) {
	folder := filepath.Base(filepath.Dir(file))

	for kind, name := range kindFolders {
		if name == folder {
			return kind, true
		}
	}

	return "", false
}

func (s Store) Lookup(
	ctx context.Context,
	id kernel.RecordID,
) (string, error) {
	prefix := id.String() + "-"

	for _, kind := range RecordKinds() {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		files, _ := listFiles(s.folder(kindFolders[kind]), extensionOf(kind))
		for _, file := range files {
			if strings.HasPrefix(filepath.Base(file), prefix) {
				return file, nil
			}
		}
	}

	details := map[string]any{"id": id.String()}

	return "", recordErrors.Fail(RecordMissing, details)
}

func extensionOf(kind string) string {
	if kind == guideKind {
		return guideExtension
	}

	return recordExtension
}

func ReadDocument(file string, data []byte) (Document, error) {
	kind, found := KindOfFile(file)
	if !found {
		return Document{}, failAt(RecordFileExtension, file, 0)
	}

	if kind != guideKind {
		root, err := ParseRecord(file, data)
		if err != nil {
			return Document{}, err
		}

		return Document{Kind: kind, File: file, JSON: objectJSON(root)}, nil
	}

	parsed, err := ParseGuide(file, data)
	if err != nil {
		return Document{}, err
	}

	match := guideFilePattern.FindStringSubmatch(filepath.Base(file))
	if match == nil {
		return Document{}, failAt(RecordFileMismatch, file, 1)
	}

	var buffer bytes.Buffer
	buffer.WriteString(`{"id":`)
	writeString(&buffer, match[1])
	writePairs(&buffer, parsed.Root, true)
	buffer.WriteString(`,"body":`)
	writeString(&buffer, parsed.Body)
	buffer.WriteByte('}')

	return Document{Kind: kind, File: file, JSON: buffer.Bytes()}, nil
}

func objectJSON(root *yaml.Node) []byte {
	var buffer bytes.Buffer
	writeNode(&buffer, root)

	return buffer.Bytes()
}

func writeNode(buffer *bytes.Buffer, node *yaml.Node) {
	switch node.Kind {
	case yaml.MappingNode:
		buffer.WriteByte('{')
		writePairs(buffer, node, false)
		buffer.WriteByte('}')
	case yaml.SequenceNode:
		buffer.WriteByte('[')
		for index, item := range node.Content {
			if index > 0 {
				buffer.WriteByte(',')
			}
			writeNode(buffer, item)
		}
		buffer.WriteByte(']')
	default:
		writeScalar(buffer, node)
	}
}

func writePairs(buffer *bytes.Buffer, mapping *yaml.Node, after bool) {
	for index := 0; index < len(mapping.Content); index += 2 {
		if after || index > 0 {
			buffer.WriteByte(',')
		}

		writeString(buffer, mapping.Content[index].Value)
		buffer.WriteByte(':')
		writeNode(buffer, mapping.Content[index+1])
	}
}

func writeScalar(buffer *bytes.Buffer, node *yaml.Node) {
	if node.Style&quotedStyles != 0 {
		writeString(buffer, node.Value)
		return
	}

	switch node.ShortTag() {
	case "!!null":
		buffer.WriteString("null")
	case "!!bool":
		buffer.WriteString(strings.ToLower(node.Value))
	case "!!int":
		writeInteger(buffer, node.Value)
	default:
		writeString(buffer, node.Value)
	}
}

func writeInteger(buffer *bytes.Buffer, value string) {
	if decimalPattern.MatchString(value) {
		buffer.WriteString(value)
		return
	}

	writeString(buffer, value)
}

func writeString(buffer *bytes.Buffer, value string) {
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(value)

	buffer.Write(bytes.TrimSuffix(encoded.Bytes(), []byte{'\n'}))
}
