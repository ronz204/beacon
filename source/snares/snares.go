package snares

import (
	"context"
	"io"

	"github.com/ronz204/beacon/source/snares/internal/action"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/system"
	"github.com/ronz204/beacon/source/weaver"
)

type Snares struct {
	ports ports
}

type ports struct {
	weaver.Weaver
	project spider.Project
}

func (p ports) Doctor(ctx context.Context) []spider.Finding {
	return p.project.Doctor(ctx)
}

func (p ports) HasHead(ctx context.Context) bool {
	_, err := system.Git(
		ctx,
		p.project.Worktree(),
		"rev-parse", "--verify", "--quiet", "HEAD",
	)

	return err == nil
}

func New(project spider.Project) Snares {
	return Snares{ports: ports{Weaver: weaver.New(project), project: project}}
}

func (s Snares) Streaming(out io.Writer) Snares {
	s.ports.Weaver = s.ports.Weaver.Streaming(out)

	return s
}

func (s Snares) PreCommit(ctx context.Context) (PreCommitView, error) {
	result, err := action.PreCommit{Ports: s.ports}.Run(ctx)

	return PreCommitView{
		Blocks:   blockViews(result.Blocks),
		Warnings: nonNil(result.Warnings),
		Checked:  nonNil(result.Checked),
	}, err
}

func blockViews(blocks []action.Block) []BlockView {
	views := make([]BlockView, 0, len(blocks))

	for _, block := range blocks {
		views = append(views, BlockView(block))
	}

	return views
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}

	return values
}
