package ledger

import "context"

type ChangeReader interface {
	FilesChangedSince(ctx context.Context, commit string) ([]string, error)
	FootprintOf(ctx context.Context, files []string) (string, error)
}
