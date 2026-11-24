package agenda_test

import (
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
)

func sampleTaskInput() agenda.TaskInput {
	return agenda.TaskInput{
		ID:     must(agenda.ParseTaskID("bc-91cd.2")),
		Author: must(kernel.ParseActor("agent:sample-agent")),
		Title:  must(kernel.ParseRecordText("Create the new schema")),
		Parent: must(agenda.ParseTaskWorkID("bc-91cd")),
		Blocks: []agenda.TaskID{
			must(agenda.ParseTaskID("bc-91cd.3")),
		},
		Implements: must(agenda.ParseTaskPromiseID("bc-c3d4")),
	}
}

func sampleTaskState(status agenda.TaskStatus) agenda.TaskState {
	input := sampleTaskInput()

	return agenda.TaskState{
		ID:      input.ID,
		Created: must(kernel.ParseCalendarDate("2026-11-22")),
		Author:  input.Author,
		Title:   input.Title,
		Parent:  input.Parent,
		Status:  status,
	}
}

func TestTaskReferenceIDs(t *testing.T) {
	_, err := agenda.ParseTaskID("bc-91cd")
	requireCode(t, err, agenda.TaskIDInvalid)

	_, err = agenda.ParseTaskWorkID("bc-91cd.1")
	requireCode(t, err, agenda.TaskParentInvalid)

	_, err = agenda.ParseTaskPromiseID("bc-c3d4.1")
	requireCode(t, err, agenda.TaskPromiseInvalid)

	nested := must(agenda.ParseTaskID("bc-91cd.1.1"))
	if nested.String() != "bc-91cd.1.1" {
		t.Fatalf("nested = %q", nested)
	}
}

func TestParseTaskStatus(t *testing.T) {
	for _, status := range agenda.TaskStatuses() {
		got, err := agenda.ParseTaskStatus(string(status))

		if err != nil {
			t.Fatal(err)
		}
		if got != status {
			t.Fatalf("got %q", got)
		}
	}

	_, err := agenda.ParseTaskStatus("blocked")
	requireCode(t, err, agenda.TaskStatusInvalid)
}

func TestParseTaskBranch(t *testing.T) {
	got, err := agenda.ParseTaskBranch(" feat/sample-branch ")

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "feat/sample-branch" {
		t.Fatalf("got %q", got)
	}

	invalid := []string{
		"",
		"two words",
		"-option",
		"a\x00b",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := agenda.ParseTaskBranch(raw)

			requireCode(t, err, agenda.TaskBranchInvalid)
		})
	}
}

func TestNewTaskStartsOpenAndDated(t *testing.T) {
	task, err := agenda.NewTask(sampleTaskInput(), sampleNow)

	if err != nil {
		t.Fatal(err)
	}
	if task.Status() != agenda.TaskOpen {
		t.Fatalf("status = %q", task.Status())
	}
	if task.Created().String() != "2026-11-22" {
		t.Fatalf("created = %q", task.Created())
	}
	if _, found := task.Owner(); found {
		t.Fatal("a new task has no owner")
	}
}

func TestNewTaskRejectsMissingField(t *testing.T) {
	input := sampleTaskInput()
	input.Title = kernel.RecordText{}

	_, err := agenda.NewTask(input, sampleNow)

	requireCode(t, err, agenda.TaskFieldMissing)
}

func TestTaskIDMustDescendFromParent(t *testing.T) {
	input := sampleTaskInput()
	input.Parent = must(agenda.ParseTaskWorkID("bc-ffff"))

	_, err := agenda.NewTask(input, sampleNow)

	requireCode(t, err, agenda.TaskParentMismatch)
}

func TestNestedTaskDescendsFromRootWork(t *testing.T) {
	input := sampleTaskInput()
	input.ID = must(agenda.ParseTaskID("bc-91cd.2.1"))

	if _, err := agenda.NewTask(input, sampleNow); err != nil {
		t.Fatal(err)
	}
}

func TestClaimedTaskNeedsOwner(t *testing.T) {
	state := sampleTaskState(agenda.TaskClaimed)

	err := agenda.ValidateTaskState(state)
	requireCode(t, err, agenda.TaskOwnerMissing)

	state.Owner = must(kernel.ParseActor("agent:sample-agent"))
	state.Branch = must(agenda.ParseTaskBranch("feat/sample-branch"))

	if err := agenda.ValidateTaskState(state); err != nil {
		t.Fatal(err)
	}

	task := agenda.RestoreTask(state)
	if _, found := task.Owner(); !found {
		t.Fatal("owner should be kept")
	}
	if _, found := task.Branch(); !found {
		t.Fatal("branch should be kept")
	}
}

func TestTaskProtectsItsBlocks(t *testing.T) {
	input := sampleTaskInput()
	task := must(agenda.NewTask(input, sampleNow))
	replacement := must(agenda.ParseTaskID("bc-91cd.9"))

	input.Blocks[0] = replacement
	task.Blocks()[0] = replacement

	if task.Blocks()[0].String() != "bc-91cd.3" {
		t.Fatalf("blocks = %v", task.Blocks())
	}
}

func TestTaskIdentityAndOptionalFields(t *testing.T) {
	task := must(agenda.NewTask(sampleTaskInput(), sampleNow))
	restored := agenda.RestoreTask(agenda.TaskState{ID: task.ID()})

	if !restored.SameAs(task) {
		t.Fatal("tasks with the same id should be the same")
	}
	if _, found := task.Implements(); !found {
		t.Fatal("implements should be present")
	}
	if _, found := restored.Implements(); found {
		t.Fatal("implements should be absent")
	}
	if _, found := task.Approved(); found {
		t.Fatal("approved should be absent")
	}
}
