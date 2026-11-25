package agenda

import "context"

type WorkRepository interface {
	ListWorks(ctx context.Context) ([]Work, []error)
}
