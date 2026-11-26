package spider

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/system"
)

type BranchView struct {
	Name     string `json:"name"`
	Detached bool   `json:"detached"`
}

func (p Project) Branch(ctx context.Context) (BranchView, error) {
	symbolic := []string{"symbolic-ref", "--short", "-q", "HEAD"}

	name, err := system.Git(ctx, p.root, symbolic...)
	if err == nil {
		return BranchView{Name: strings.TrimSpace(name)}, nil
	}

	commit, err := system.Git(ctx, p.root, "rev-parse", "HEAD")
	if err != nil {
		return BranchView{}, kernel.FailHistory(err)
	}

	return BranchView{Name: strings.TrimSpace(commit), Detached: true}, nil
}

func (p Project) RepoPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.ToSlash(path)
	}

	return p.relative(absolute)
}
