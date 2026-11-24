package memory

import (
	"context"

	"github.com/ronz204/beacon/internal/kernel"
)

type CodeCommit struct {
	Day   kernel.CalendarDate
	Files []string
}

type HistoryReader interface {
	CommitsSince(
		ctx context.Context,
		day kernel.CalendarDate,
	) ([]CodeCommit, error)
	GuideChangeDays(
		ctx context.Context,
	) (map[GuideID]kernel.CalendarDate, error)
}
