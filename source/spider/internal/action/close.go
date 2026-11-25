package action

import (
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
)

func PromiseOwner(snapshot Snapshot, promise string) (string, bool) {
	for _, capability := range snapshot.Capabilities {
		for _, item := range capability.Promises() {
			if item.ID().String() == promise {
				return capability.ID().String(), true
			}
		}
	}

	return "", false
}

func OpenSubtasks(snapshot Snapshot, task string) []string {
	prefix := task + "."
	var open []string

	for _, candidate := range snapshot.Tasks {
		id := candidate.ID().String()
		if !strings.HasPrefix(id, prefix) {
			continue
		}

		if candidate.Status() != agenda.TaskDone {
			open = append(open, id)
		}
	}
	slices.Sort(open)

	return open
}
