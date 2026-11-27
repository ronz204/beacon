package spider

import (
	"context"
	"encoding/json"

	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
)

var searchedFields = []string{
	"statement", "success", "name", "intent", "text",
	"what", "why", "request", "title",
}

func (p Project) Find(
	ctx context.Context,
	text string,
	archived bool,
) (FindView, []Finding) {
	var candidates []action.FindCandidate
	var failures []error

	if archived {
		var err error
		candidates, err = p.archivedCandidates(ctx)
		if err != nil {
			failures = append(failures, err)
		}
	}

	result, findFailures := action.Find(
		ctx,
		p.repositories(),
		text,
		candidates,
	)

	return findView(result), p.findings(failures, findFailures)
}

func (p Project) archivedCandidates(
	ctx context.Context,
) ([]action.FindCandidate, error) {
	entries, err := p.archive.Entries(ctx)
	if err != nil {
		return nil, err
	}

	candidates := make([]action.FindCandidate, 0, len(entries))
	for _, entry := range entries {
		var fields map[string]any
		if json.Unmarshal(entry.Record, &fields) != nil {
			continue
		}

		id, _ := fields["id"].(string)
		candidate := action.FindCandidate{
			Kind: thread.NodeKind(entry.Type),
			ID:   id,
		}
		for _, key := range searchedFields {
			if text, isText := fields[key].(string); isText {
				candidate.Texts = append(candidate.Texts, text)
			}
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}
