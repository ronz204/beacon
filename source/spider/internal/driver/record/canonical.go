package record

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
	"unicode"

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
	configKeys = []string{
		"schema", "runners", "review", "limits", "agents", "approved",
	}
	reviewKeys = []string{
		"convention", "architecture", "product", "domain", "runbook",
		"decision",
	}
	limitKeys    = []string{"guide", "brief", "hook"}
	runnerKeys   = []string{"run", "baseline"}
	approvedKeys = []string{"by", "at", "via", "n"}
	attestedKeys = []string{"by", "at", "via", "commit", "evidence"}
	checkKeys    = []string{"runner", "manual"}
	flowLists    = []string{
		"paths", "where", "blocks", "implements", "refers", "tags", "applies",
		"run", "baseline", "agents",
	}
	sortedLists = []string{"blocks", "implements", "refers"}
	blockLists  = []string{"nongoals", "rejected", "log"}
	typedFields = []string{
		"created", "reviewed", "at", "always", "schema", "guide", "brief",
		"hook", "n",
	}
)

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
	case "runners":
		section.Style = 0
		canonicalRunners(section)
	case "review":
		section.Style = 0
		canonicalMapping(section, reviewKeys)
	case "limits":
		section.Style = 0
		canonicalMapping(section, limitKeys)
	}
}

func canonicalRunners(section *yaml.Node) {
	orderKeys(section, nil)

	for index := 1; index < len(section.Content); index += 2 {
		if runner := section.Content[index]; runner.Kind == yaml.MappingNode {
			runner.Style = 0
			canonicalMapping(runner, runnerKeys)
		}
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
