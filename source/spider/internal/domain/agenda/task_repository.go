package agenda

import "context"

type TaskRepository interface {
	ListTasks(ctx context.Context) ([]Task, []error)
}
