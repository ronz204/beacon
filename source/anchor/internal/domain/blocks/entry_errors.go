package blocks

import (
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

type EntryCode string

const (
	EntryInvalid    EntryCode = "entry_invalid"
	EntryConflict   EntryCode = "entry_conflict"
	DocumentInvalid EntryCode = "document_invalid"
)

type EntryError = kernel.Error[EntryCode]

var entryErrors = kernel.ErrorCatalog[EntryCode]{
	EntryInvalid:    "an entry needs a known mode, a key path and a JSON value",
	EntryConflict:   "the file already holds a different value at that path",
	DocumentInvalid: "the file must be a JSON object",
}

func invalidEntry(path []string) EntryError {
	details := map[string]any{"path": strings.Join(path, ".")}
	return entryErrors.Fail(EntryInvalid, details)
}

func conflictingEntry(path []string) EntryError {
	details := map[string]any{"path": strings.Join(path, ".")}
	return entryErrors.Fail(EntryConflict, details)
}

func invalidDocument(cause error) EntryError {
	return entryErrors.Fail(DocumentInvalid, nil).WithCause(cause)
}
