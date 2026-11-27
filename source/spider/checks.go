package spider

import (
	"context"
	"errors"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
	"github.com/ronz204/beacon/source/system"
)

var errUnsafeRevision = errors.New("a revision cannot start with a dash")

type CheckPromise struct {
	ID     string            `json:"id"`
	Status string            `json:"status"`
	Runner string            `json:"runner,omitempty"`
	Params map[string]string `json:"params,omitempty"`
	Manual string            `json:"manual,omitempty"`
}

type CheckTarget struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Paths    []string       `json:"paths"`
	Promises []CheckPromise `json:"promises"`
}

func (p Project) CheckTargets(ctx context.Context) ([]CheckTarget, []Finding) {
	snapshot := action.Load(ctx, p.repositories())
	targets := make([]CheckTarget, 0, len(snapshot.Capabilities))

	for _, capability := range snapshot.Capabilities {
		target := CheckTarget{
			ID:       capability.ID().String(),
			Name:     capability.Name().String(),
			Paths:    globTexts(capability.Paths()),
			Promises: []CheckPromise{},
		}

		for _, promise := range capability.Promises() {
			check := promise.Check()
			target.Promises = append(target.Promises, CheckPromise{
				ID:     promise.ID().String(),
				Status: string(promise.Status()),
				Runner: check.Runner(),
				Params: check.Params(),
				Manual: check.Manual(),
			})
		}

		targets = append(targets, target)
	}

	return targets, p.findings(snapshot.Failures)
}

func (p Project) TouchedCapabilities(
	ctx context.Context,
	revision string,
) ([]string, error) {
	resolved, err := p.resolveRevision(ctx, revision)
	if err != nil {
		return nil, err
	}

	files, err := commit.NewRepository(p.root).FilesChangedSince(ctx, resolved)
	if err != nil {
		return nil, err
	}

	snapshot := action.Load(ctx, p.repositories())
	matcher := match.NewMatcher()
	var touched []string

	for _, capability := range snapshot.Capabilities {
		isTouched, err := touches(matcher, capability.Paths(), files)
		if err != nil {
			return nil, err
		}
		if isTouched {
			touched = append(touched, capability.ID().String())
		}
	}

	return touched, nil
}

func touches(
	matcher kernel.PathMatcher,
	globs []kernel.PathGlob,
	files []string,
) (bool, error) {
	for _, glob := range globs {
		for _, file := range files {
			matched, err := matcher.MatchPath(glob, file)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
	}

	return false, nil
}

func globTexts(globs []kernel.PathGlob) []string {
	texts := make([]string, 0, len(globs))

	for _, glob := range globs {
		texts = append(texts, glob.String())
	}

	return texts
}

func (p Project) resolveRevision(
	ctx context.Context,
	revision string,
) (string, error) {
	if strings.HasPrefix(revision, "-") {
		return "", kernel.FailHistory(errUnsafeRevision)
	}

	output, err := system.Git(
		ctx,
		p.root,
		"rev-parse", "--verify", "--quiet", revision+"^{commit}",
	)
	if err != nil {
		return "", kernel.FailHistory(err)
	}

	return strings.TrimSpace(output), nil
}
