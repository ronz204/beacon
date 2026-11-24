package agenda

import "github.com/ronz204/beacon/internal/kernel"

func ValidateTaskState(state TaskState) error {
	field := missingTaskField(state)
	if field != "" {
		details := map[string]any{"field": field}
		return taskErrors.Fail(TaskFieldMissing, details)
	}

	if state.ID.value.Root() != state.Parent.value {
		return invalidTask(TaskParentMismatch, state.ID.String())
	}
	if state.Status == TaskClaimed && kernel.IsMissing(state.Owner) {
		return invalidTask(TaskOwnerMissing, state.ID.String())
	}

	return nil
}

func missingTaskField(state TaskState) string {
	switch {
	case kernel.IsMissing(state.ID):
		return "id"
	case kernel.IsMissing(state.Created):
		return "created"
	case kernel.IsMissing(state.Author):
		return "author"
	case kernel.IsMissing(state.Title):
		return "title"
	case kernel.IsMissing(state.Parent):
		return "parent"
	case kernel.IsMissing(state.Status):
		return "status"
	}

	return ""
}
