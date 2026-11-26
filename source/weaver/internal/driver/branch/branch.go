package branch

import (
	"context"
	"strings"

	"github.com/ronz204/beacon/source/system"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

var defaultBases = []string{"origin/HEAD", "main", "master"}

type Git struct {
	Directory string
}

func (g Git) Base(ctx context.Context, forced string) (string, error) {
	candidates := defaultBases
	if forced != "" {
		candidates = []string{forced}
	}
	if strings.HasPrefix(forced, "-") {
		return "", action.FailBaseMissing(forced)
	}

	for _, candidate := range candidates {
		output, err := system.Git(
			ctx,
			g.Directory,
			"merge-base", "HEAD", candidate,
		)
		if err == nil {
			return strings.TrimSpace(output), nil
		}
	}

	return "", action.FailBaseMissing(forced)
}
