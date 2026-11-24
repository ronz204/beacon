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
	RecordFieldMissing       RecordCode = "record_field_missing"
	RecordFieldType          RecordCode = "record_field_type"
	RecordFileMismatch       RecordCode = "record_file_mismatch"
	RecordFileExtension      RecordCode = "record_file_extension"
	RecordSchemaMissing      RecordCode = "record_schema_missing"
	RecordSchemaNewer        RecordCode = "record_schema_newer"
	RecordFormatUnsafe       RecordCode = "record_format_unsafe"
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
	RecordFieldMissing:       "add the required field",
	RecordFieldType:          "use the type the schema defines for this field",
	RecordFileMismatch:       "rename the file to start with the record id",
	RecordFileExtension:      "use the extension this record folder expects",
	RecordSchemaMissing:      "add schema: 1 to config.yaml",
	RecordSchemaNewer:        "update beacon to read this schema version",
	RecordFormatUnsafe:       "formatting would alter the record content",
}

func failAt(code RecordCode, file string, line int) RecordError {
	details := map[string]any{
		"file": file,
		"line": line,
	}

	return recordErrors.Fail(code, details)
}
