package ledger

import "context"

type ChangeReader interface {
	FilesChangedSince(ctx context.Context, commit string) ([]string, error)
}
