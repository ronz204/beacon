package record

import (
	"bytes"
	"cmp"
	"context"
	"os"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/system"
	yaml "go.yaml.in/yaml/v4"
)

const (
	frontmatterEnd = "---\n"
	stringTag      = "!!str"
	yamlIndicators = "*&!%@`?|>[{'\",]}"
	flowCharacters = ",{}[]"

	maximumRenderPasses = 3
)

var (
	numberPattern = regexp.MustCompile(
		`(?i)^[-+]?(\.inf|\.nan|0x[0-9a-f_]+|0o[0-7_]+|0b[01_]+|` +
			`[0-9][0-9_]*(\.[0-9_]*)?(e[-+]?[0-9]+)?|` +
			`\.[0-9_]+(e[-+]?[0-9]+)?)$`,
	)
	datePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{1,2}-[0-9]{1,2}$`)
)

var reservedWords = []string{
	"y",
	"n",
	"yes",
	"no",
	"on",
	"off",
	"true",
	"false",
	"null",
	"~",
}

var recordKeyOrders = map[string][]string{
	"goals": {
		"id", "created", "author", "statement", "success",
		"nongoals", "status", "approved",
	},
	"capabilities": {
		"id", "created", "author", "name", "intent", "paths",
		"serves", "nongoals", "approved", "promises",
	},
	"decisions": {
		"id", "created", "author", "what", "why", "where", "supersedes",
		"rejected", "status", "review", "reviewed", "approved",
	},
	"questions": {
		"id", "created", "author", "text", "blocks", "status",
		"answer", "origin", "approved",
	},
	"works": {
		"id", "created", "author", "request", "implements", "serves",
		"status", "log", "approved",
	},
	"tasks": {
		"id", "created", "author", "title", "parent", "blocks",
		"implements", "status", "owner", "branch", "approved",
	},
}

var (
	guideKeys = []string{
		"title", "kind", "applies", "always", "review", "reviewed",
		"tags", "audience", "status", "refers",
	}
	promiseKeys = []string{
		"id", "text", "check", "status", "approved", "attested",
		"origin", "refers",
	}
	approvedKeys = []string{"by", "at", "via"}
	attestedKeys = []string{"by", "at", "commit", "evidence"}
	checkKeys    = []string{"runner", "manual"}
	flowLists    = []string{
		"paths", "where", "blocks", "implements", "refers", "tags", "applies",
	}
	sortedLists = []string{"blocks", "implements", "refers"}
	blockLists  = []string{"nongoals", "rejected", "log"}
	typedFields = []string{"created", "reviewed", "at", "always"}
)

type formatTarget struct {
	folder    string
	extension string
	check     func(file string, data []byte) []error
	format    func(file string, data []byte) ([]byte, error)
}

func FormatRecord(file, folder string, data []byte) ([]byte, error) {
	return renderStable(file, normalizeText(data), 0, recordKeyOrders[folder])
}

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

func FormatGuide(file string, data []byte) ([]byte, error) {
	front, body, found := splitFrontmatter(normalizeText(data))
	if !found {
		return nil, failAt(RecordFrontmatterMissing, file, 1)
	}

	rendered, err := renderStable(file, front, frontmatterOffset, guideKeys)
	if err != nil {
		return nil, err
	}

	text := frontmatterOpen + string(rendered) + frontmatterEnd + body

	return []byte(text), nil
}

func (s Store) Format(ctx context.Context, write bool) ([]string, []error) {
	var changed []string
	var failures []error

	for _, target := range formatTargets() {
		files, listed := listFiles(s.folder(target.folder), target.extension)
		failures = append(failures, listed...)

		for _, file := range files {
			differs, failed := formatFile(ctx, file, target, write)
			failures = append(failures, failed...)

			if differs {
				changed = append(changed, file)
			}
		}
	}

	return changed, failures
}

func formatTargets() []formatTarget {
	return []formatTarget{
		recordTarget("goals", checkedBy(decodeGoal)),
		recordTarget("capabilities", checkedBy(decodeCapability)),
		recordTarget("decisions", checkedBy(decodeDecision)),
		recordTarget("questions", checkedBy(decodeQuestion)),
		recordTarget("works", checkedBy(decodeWork)),
		recordTarget("tasks", checkedBy(decodeTask)),
		{
			folder:    "guides",
			extension: guideExtension,
			check:     checkedBy(decodeGuide),
			format:    FormatGuide,
		},
	}
}

func recordTarget(
	folder string,
	check func(string, []byte) []error,
) formatTarget {
	return formatTarget{
		folder:    folder,
		extension: recordExtension,
		check:     check,
		format: func(file string, data []byte) ([]byte, error) {
			return FormatRecord(file, folder, data)
		},
	}
}

func checkedBy[T any](decode decoder[T]) func(string, []byte) []error {
	return func(file string, data []byte) []error {
		_, failures := decode(file, data)
		return failures
	}
}

func formatFile(
	ctx context.Context,
	file string,
	target formatTarget,
	write bool,
) (bool, []error) {
	if err := ctx.Err(); err != nil {
		return false, []error{kernel.NewRecordFailure(file, 0, err)}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return false, []error{kernel.NewRecordFailure(file, 0, err)}
	}

	if failures := target.check(file, data); len(failures) > 0 {
		return false, failures
	}

	output, err := target.format(file, data)
	if err != nil {
		return false, parseFailure(file, err)
	}
	if bytes.Equal(output, data) {
		return false, nil
	}
	if !write {
		return true, nil
	}

	if err := system.WriteFileAtomic(file, output); err != nil {
		return false, []error{kernel.NewRecordFailure(file, 0, err)}
	}

	return true, nil
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

func canonicalMapping(mapping *yaml.Node, order []string) {
	orderKeys(mapping, order)

	for index := 0; index < len(mapping.Content); index += 2 {
		key := mapping.Content[index]
		if !slices.Contains(order, key.Value) {
			continue
		}

		canonicalPlain(key)
		canonicalValue(
			key.Value,
			mapping.Content[index+1],
			mapping.Style&yaml.FlowStyle != 0,
		)
	}
}

func orderKeys(mapping *yaml.Node, order []string) {
	entries := make([][2]*yaml.Node, 0, len(mapping.Content)/2)
	for index := 0; index < len(mapping.Content); index += 2 {
		entries = append(entries, [2]*yaml.Node{
			mapping.Content[index],
			mapping.Content[index+1],
		})
	}

	slices.SortStableFunc(entries, func(a, b [2]*yaml.Node) int {
		return compareKeys(a[0].Value, b[0].Value, order)
	})

	mapping.Content = mapping.Content[:0]
	for _, entry := range entries {
		mapping.Content = append(mapping.Content, entry[0], entry[1])
	}
}

func compareKeys(a, b string, order []string) int {
	rankA := slices.Index(order, a)
	rankB := slices.Index(order, b)

	switch {
	case rankA >= 0 && rankB >= 0:
		return cmp.Compare(rankA, rankB)
	case rankA >= 0:
		return -1
	case rankB >= 0:
		return 1
	}

	return cmp.Compare(a, b)
}

func canonicalValue(name string, value *yaml.Node, flow bool) {
	switch value.Kind {
	case yaml.ScalarNode:
		canonicalScalar(value, slices.Contains(typedFields, name), flow)
	case yaml.SequenceNode:
		canonicalSequence(name, value)
	case yaml.MappingNode:
		canonicalSection(name, value)
	}
}

func canonicalSequence(name string, sequence *yaml.Node) {
	switch {
	case name == "promises":
		canonicalPromises(sequence)
	case slices.Contains(flowLists, name):
		sequence.Style = inlineStyle(sequence)
		canonicalItems(sequence, true)
		if slices.Contains(sortedLists, name) {
			sortByValue(sequence)
		}
	case slices.Contains(blockLists, name):
		sequence.Style = 0
		canonicalItems(sequence, false)
	}
}

func canonicalItems(sequence *yaml.Node, flow bool) {
	for _, item := range sequence.Content {
		if item.Kind == yaml.ScalarNode {
			canonicalScalar(item, false, flow)
		}
	}
}

func sortByValue(sequence *yaml.Node) {
	slices.SortStableFunc(sequence.Content, func(a, b *yaml.Node) int {
		return cmp.Compare(a.Value, b.Value)
	})
}

func canonicalPromises(sequence *yaml.Node) {
	sequence.Style = 0

	for _, item := range sequence.Content {
		if item.Kind == yaml.MappingNode {
			item.Style = 0
			canonicalMapping(item, promiseKeys)
		}
	}

	slices.SortStableFunc(sequence.Content, func(a, b *yaml.Node) int {
		return cmp.Compare(promiseID(a), promiseID(b))
	})
}

func promiseID(item *yaml.Node) string {
	for index := 0; index+1 < len(item.Content); index += 2 {
		if item.Content[index].Value == "id" {
			return item.Content[index+1].Value
		}
	}

	return ""
}

func canonicalSection(name string, section *yaml.Node) {
	switch name {
	case "check":
		section.Style = inlineStyle(section)
		canonicalCheck(section)
	case "approved":
		section.Style = inlineStyle(section)
		canonicalMapping(section, approvedKeys)
	case "attested":
		section.Style = inlineStyle(section)
		canonicalMapping(section, attestedKeys)
	}
}

func canonicalCheck(section *yaml.Node) {
	orderKeys(section, checkKeys)

	for index := 0; index < len(section.Content); index += 2 {
		canonicalPlain(section.Content[index])

		if value := section.Content[index+1]; value.Kind == yaml.ScalarNode {
			canonicalScalar(value, false, true)
		}
	}
}

func canonicalScalar(node *yaml.Node, typed, flow bool) {
	switch {
	case typed:
		canonicalPlain(node)
	case strings.Contains(node.Value, "\n") && !flow:
		node.Style = yaml.LiteralStyle
		node.Tag = stringTag
	case needsQuotes(node.Value, flow):
		node.Style = yaml.DoubleQuotedStyle
		node.Tag = stringTag
	default:
		canonicalPlain(node)
	}
}

func canonicalPlain(node *yaml.Node) {
	node.Style = 0
	node.Tag = ""
}

func needsQuotes(value string, flow bool) bool {
	switch {
	case value == "", value != strings.TrimSpace(value):
		return true
	case slices.Contains(reservedWords, strings.ToLower(value)):
		return true
	case numberPattern.MatchString(value), datePattern.MatchString(value):
		return true
	case strings.ContainsAny(value, ":#"), startsWithIndicator(value):
		return true
	case strings.ContainsFunc(value, unicode.IsControl):
		return true
	}

	return flow && strings.ContainsAny(value, flowCharacters)
}

func startsWithIndicator(value string) bool {
	if value == "-" || strings.HasPrefix(value, "- ") {
		return true
	}

	return strings.ContainsRune(yamlIndicators, rune(value[0]))
}
