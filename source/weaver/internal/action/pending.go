package action

import (
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

type Entry struct {
	Path     string
	Kind     string
	Finding  deltas.Finding
	Approved bool
	Runners  []string
}

func EntriesOf(changes []deltas.StateChange) []Entry {
	var entries []Entry

	for _, change := range changes {
		for _, finding := range deltas.ClassifyElements(change) {
			entries = append(entries, Entry{
				Path:    change.Path(),
				Kind:    string(change.Kind()),
				Finding: finding,
				Runners: change.Runners(),
				Approved: finding.NeedsYes &&
					deltas.IsFreshlyApproved(finding.Target, changes),
			})
		}
	}

	slices.SortStableFunc(entries, func(a, b Entry) int {
		if a.Finding.Level != b.Finding.Level {
			return int(b.Finding.Level) - int(a.Finding.Level)
		}

		return strings.Compare(a.Path, b.Path)
	})

	return entries
}

func PendingOf(entries []Entry) []Entry {
	var pending []Entry

	for _, entry := range entries {
		if entry.Finding.NeedsYes && !entry.Approved {
			pending = append(pending, entry)
		}
	}

	return pending
}
