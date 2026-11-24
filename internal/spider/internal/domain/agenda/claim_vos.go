package agenda

import "github.com/ronz204/beacon/internal/kernel"

type TaskClaim struct {
	Task     TaskID
	Branch   TaskBranch
	Worktree string
	Day      kernel.CalendarDate
}
