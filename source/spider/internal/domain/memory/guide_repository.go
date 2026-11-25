package memory

import "context"

type GuideRepository interface {
	ListGuides(ctx context.Context) ([]Guide, []error)
}
