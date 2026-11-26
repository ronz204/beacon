package ledger

import (
	"context"
	"errors"
)

var ErrCommitUnknown = errors.New("the commit is not in this clone")

type ChangeReader interface {
	FilesChangedSince(ctx context.Context, commit string) ([]string, error)
	FootprintOf(ctx context.Context, files []string) (string, error)
}
