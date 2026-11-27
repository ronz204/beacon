package record

import (
	"bytes"
	"slices"

	yaml "go.yaml.in/yaml/v4"
)

func renderStable(
	file string,
	text string,
	offset int,
	order []string,
) ([]byte, error) {
	current := text

	for pass := range maximumRenderPasses {
		document, err := parseDocument(file, current, offset)
		if err != nil && pass > 0 {
			return nil, failAt(RecordFormatUnsafe, file, 1).WithCause(err)
		}
		if err != nil {
			return nil, err
		}

		output, err := renderDocument(file, document, order)
		if err != nil {
			return nil, err
		}
		if string(output) == current {
			return output, nil
		}

		current = string(output)
		offset = 0
	}

	return nil, failAt(RecordFormatUnsafe, file, 1)
}

func renderDocument(
	file string,
	document *yaml.Node,
	order []string,
) ([]byte, error) {
	before := contentLines(document)
	root := document.Content[0]
	root.Style = 0
	canonicalMapping(root, order)

	output, err := encodeDocument(document)
	if err != nil {
		return nil, failAt(RecordFormatUnsafe, file, 1).WithCause(err)
	}

	reread, err := parseDocument(file, string(output), 0)
	if err != nil {
		return nil, failAt(RecordFormatUnsafe, file, 1).WithCause(err)
	}
	if !slices.Equal(before, contentLines(reread)) {
		return nil, failAt(RecordFormatUnsafe, file, 1)
	}

	return output, nil
}

func encodeDocument(document *yaml.Node) ([]byte, error) {
	var buffer bytes.Buffer

	dumper, err := yaml.NewDumper(
		&buffer,
		yaml.WithV3Defaults(),
		yaml.WithIndent(2),
		yaml.WithLineWidth(-1),
	)
	if err != nil {
		return nil, err
	}

	if err := dumper.Dump(document); err != nil {
		return nil, err
	}
	if err := dumper.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func inlineStyle(node *yaml.Node) yaml.Style {
	if hasComments(node) {
		return 0
	}

	return yaml.FlowStyle
}

func hasComments(node *yaml.Node) bool {
	for _, child := range node.Content {
		if child.HeadComment != "" || child.LineComment != "" {
			return true
		}
		if child.FootComment != "" || hasComments(child) {
			return true
		}
	}

	return false
}

func contentLines(document *yaml.Node) []string {
	var lines []string

	collectContent(document.Content[0], "", &lines)
	slices.Sort(lines)

	return lines
}

func collectContent(node *yaml.Node, path string, lines *[]string) {
	switch node.Kind {
	case yaml.MappingNode:
		for index := 0; index < len(node.Content); index += 2 {
			child := path + "/" + node.Content[index].Value
			*lines = append(*lines, child)
			collectContent(node.Content[index+1], child, lines)
		}
	case yaml.SequenceNode:
		for _, item := range node.Content {
			collectContent(item, path+"[]", lines)
		}
	default:
		*lines = append(*lines, path+"="+node.Value)
	}
}
