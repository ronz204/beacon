package kernel

import "context"

type RecordFormatter interface {
	Format(ctx context.Context, write bool) ([]string, []error)
}
