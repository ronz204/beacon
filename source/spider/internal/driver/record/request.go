package record

import (
	"slices"

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
	_, known := kindFolders[kindNode.Value]
	if !known && kindNode.Value != ConfigKind {
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
