package action

import "github.com/ronz204/beacon/source/kernel"

type InstallCode string

const (
	FileReadFailed   InstallCode = "file_read_failed"
	FileWriteFailed  InstallCode = "file_write_failed"
	FileRemoveFailed InstallCode = "file_remove_failed"
	BackupFailed     InstallCode = "backup_failed"
	FilePathInvalid  InstallCode = "file_path_invalid"
	PackUnknown      InstallCode = "pack_unknown"
	PieceUnknown     InstallCode = "piece_unknown"
	PackAssetInvalid InstallCode = "pack_asset_invalid"
)

type InstallError = kernel.Error[InstallCode]

var installErrors = kernel.ErrorCatalog[InstallCode]{
	FileReadFailed:   "a file could not be read",
	FileWriteFailed:  "a file could not be written",
	FileRemoveFailed: "a file could not be removed",
	BackupFailed:     "a backup copy could not be saved",
	FilePathInvalid:  "a file path must stay inside the repository",
	PackUnknown:      "no pack has that name",
	PieceUnknown:     "no rule or skill has that name",
	PackAssetInvalid: "a pack file could not be read or rendered",
}

func FailInstall(
	code InstallCode,
	path string,
	cause error,
) InstallError {
	failure := installErrors.Fail(code, map[string]any{"path": path})

	if cause == nil {
		return failure
	}

	return failure.WithCause(cause)
}

func FailPackName(code InstallCode, name string) InstallError {
	return installErrors.Fail(code, map[string]any{"name": name})
}
