package thread

import "context"

type ArchiveIndex interface {
	ArchivedNodes(ctx context.Context) ([]GraphNode, error)
}
