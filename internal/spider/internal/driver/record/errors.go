package record

import "github.com/ronz204/beacon/internal/kernel"

type RecordCode string

const (
	RecordSyntaxInvalid      RecordCode = "record_syntax_invalid"
	RecordAnchorForbidden    RecordCode = "record_anchor_forbidden"
	RecordMergeForbidden     RecordCode = "record_merge_forbidden"
	RecordTagForbidden       RecordCode = "record_tag_forbidden"
	RecordKeyInvalid         RecordCode = "record_key_invalid"
	RecordKeyRepeated        RecordCode = "record_key_repeated"
	RecordDocumentEmpty      RecordCode = "record_document_empty"
	RecordDocumentExtra      RecordCode = "record_document_extra"
	RecordRootInvalid        RecordCode = "record_root_invalid"
	RecordFrontmatterMissing RecordCode = "record_frontmatter_missing"
)

type RecordError = kernel.Error[RecordCode]

var recordErrors = kernel.ErrorCatalog[RecordCode]{
	RecordSyntaxInvalid:      "fix the YAML syntax at the reported line",
	RecordAnchorForbidden:    "replace anchors and aliases with full values",
	RecordMergeForbidden:     "replace << with the explicit keys",
	RecordTagForbidden:       "remove the explicit YAML tag",
	RecordKeyInvalid:         "use plain text for mapping keys",
	RecordKeyRepeated:        "keep only one copy of the repeated key",
	RecordDocumentEmpty:      "add the record fields to the empty file",
	RecordDocumentExtra:      "keep a single YAML document in the file",
	RecordRootInvalid:        "the record must be a YAML mapping",
	RecordFrontmatterMissing: "start the guide with a --- frontmatter block",
}

func failAt(code RecordCode, file string, line int) RecordError {
	details := map[string]any{
		"file": file,
		"line": line,
	}

	return recordErrors.Fail(code, details)
}
