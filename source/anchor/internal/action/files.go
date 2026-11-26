package action

import "context"

type FileStore interface {
	Read(ctx context.Context, path string) (string, bool, error)
	Write(ctx context.Context, path, content string) error
	Remove(ctx context.Context, path string) error
	Backup(ctx context.Context, stamp, path, content string) error
}
