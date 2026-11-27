package record

import (
	"errors"
	"io"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

const (
	byteOrderMark     = "\xef\xbb\xbf"
	frontmatterOpen   = "---\n"
	frontmatterClose  = "\n---\n"
	frontmatterOffset = 1
)

type ParsedGuide struct {
	Root *yaml.Node
	Body string
}

func ParseRecord(file string, data []byte) (*yaml.Node, error) {
	text := normalizeText(data)
	if line := conflictLine(text); line > 0 {
		return nil, failAt(RecordConflict, file, line)
	}

	return parseMapping(file, text, 0)
}

func ParseGuide(file string, data []byte) (ParsedGuide, error) {
	text := normalizeText(data)
	if line := conflictLine(text); line > 0 {
		return ParsedGuide{}, failAt(RecordConflict, file, line)
	}

	front, body, found := splitFrontmatter(text)
	if !found {
		return ParsedGuide{}, failAt(RecordFrontmatterMissing, file, 1)
	}

	root, err := parseMapping(file, front, frontmatterOffset)
	if err != nil {
		return ParsedGuide{}, err
	}

	return ParsedGuide{Root: root, Body: body}, nil
}

func conflictLine(text string) int {
	opened, middle := 0, 0

	for number, line := range strings.Split(text, "\n") {
		switch {
		case opened == 0 && strings.HasPrefix(line, "<<<<<<<"):
			opened = number + 1
		case opened > 0 && strings.HasPrefix(line, "======="):
			middle = number + 1
		case middle > 0 && strings.HasPrefix(line, ">>>>>>>"):
			return opened
		}
	}

	return 0
}

func normalizeText(data []byte) string {
	text := strings.TrimPrefix(string(data), byteOrderMark)
	text = strings.ReplaceAll(text, "\r\n", "\n")

	return strings.ReplaceAll(text, "\r", "\n")
}

func splitFrontmatter(text string) (string, string, bool) {
	rest, opened := strings.CutPrefix(text, frontmatterOpen)
	if !opened {
		return "", "", false
	}

	framed := "\n" + rest + "\n"
	front, body, closed := strings.Cut(framed, frontmatterClose)
	if !closed {
		return "", "", false
	}

	front = strings.TrimPrefix(front, "\n")
	body = strings.TrimSuffix(body, "\n")

	return front, body, true
}

func parseMapping(file, text string, offset int) (*yaml.Node, error) {
	document, err := parseDocument(file, text, offset)
	if err != nil {
		return nil, err
	}

	return document.Content[0], nil
}

func parseDocument(file, text string, offset int) (*yaml.Node, error) {
	decoder := yaml.NewDecoder(strings.NewReader(text))

	var document yaml.Node
	err := decoder.Decode(&document)
	if errors.Is(err, io.EOF) {
		return nil, failAt(RecordDocumentEmpty, file, offset+1)
	}
	if err != nil {
		return nil, syntaxFailure(file, err, offset)
	}

	var extra yaml.Node
	err = decoder.Decode(&extra)
	if err == nil {
		return nil, failAt(RecordDocumentExtra, file, extra.Line+offset)
	}
	if !errors.Is(err, io.EOF) {
		return nil, syntaxFailure(file, err, offset)
	}

	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, failAt(RecordRootInvalid, file, root.Line+offset)
	}
	if err := validateNode(file, root, offset); err != nil {
		return nil, err
	}

	return &document, nil
}

func syntaxFailure(file string, cause error, offset int) RecordError {
	line := 1

	if located, ok := errors.AsType[*yaml.LoadError](cause); ok {
		line = max(located.Mark.Line, 1)
	}

	return failAt(RecordSyntaxInvalid, file, line+offset).WithCause(cause)
}

func validateNode(file string, node *yaml.Node, offset int) error {
	line := node.Line + offset

	if node.Anchor != "" || node.Kind == yaml.AliasNode {
		return failAt(RecordAnchorForbidden, file, line)
	}
	if node.Style&yaml.TaggedStyle != 0 {
		return failAt(RecordTagForbidden, file, line)
	}
	if node.Kind == yaml.MappingNode {
		if err := validateKeys(file, node, offset); err != nil {
			return err
		}
	}

	for _, child := range node.Content {
		if err := validateNode(file, child, offset); err != nil {
			return err
		}
	}

	node.Line = line

	return nil
}

func validateKeys(file string, mapping *yaml.Node, offset int) error {
	seen := make(map[string]bool, len(mapping.Content)/2)

	for index := 0; index < len(mapping.Content); index += 2 {
		key := mapping.Content[index]
		line := key.Line + offset

		switch {
		case key.Kind != yaml.ScalarNode:
			return failAt(RecordKeyInvalid, file, line)
		case key.ShortTag() == "!!merge":
			return failAt(RecordMergeForbidden, file, line)
		case seen[key.Value]:
			return failAt(RecordKeyRepeated, file, line)
		}

		seen[key.Value] = true
	}

	return nil
}
