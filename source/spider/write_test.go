package spider_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
)

const (
	author         = "agent:sample-agent"
	capabilityFile = ".beacon/records/capabilities/bc-7f3a-auth-login.yaml"
)

func payloadOf(
	t *testing.T,
	kind string,
	fields map[string]any,
) []byte {
	t.Helper()

	data, err := json.Marshal(map[string]any{"type": kind, "record": fields})
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func storedFields(
	t *testing.T,
	project spider.Project,
	id string,
) map[string]any {
	t.Helper()

	stored, err := project.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]any
	if err := json.Unmarshal(stored.Data, &fields); err != nil {
		t.Fatal(err)
	}

	return fields
}

func readFile(t *testing.T, root, file string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(file)))
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func writeFile(t *testing.T, root, file, text string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(file))
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func save(
	t *testing.T,
	project spider.Project,
	payload []byte,
) spider.Prepared {
	t.Helper()

	prepared, err := project.Prepare(context.Background(), payload, author)
	if err != nil {
		t.Fatalf("prepare: %v", kernel.DescribeFailure(err))
	}

	if err := project.Save(context.Background(), prepared, nil); err != nil {
		t.Fatalf("save: %v", kernel.DescribeFailure(err))
	}

	return prepared
}

func requireCode(t *testing.T, err error, code string) kernel.FailureView {
	t.Helper()

	view := kernel.DescribeFailure(err)
	if view.Code != code {
		t.Fatalf("code = %q, want %q (%+v)", view.Code, code, view)
	}

	return view
}

func promisesOf(fields map[string]any) []any {
	promises, _ := fields["promises"].([]any)
	return promises
}

func TestGetThenRecordChangesNoByte(t *testing.T) {
	project, root := newProject(t)
	for _, id := range []string{"bc-7f3a", "bc-0a9b", "bc-91cd.1"} {
		stored, err := project.Get(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		before := readFile(t, root, stored.File)
		payload, err := json.Marshal(stored)
		if err != nil {
			t.Fatal(err)
		}

		prepared := save(t, project, payload)

		if !prepared.Unchanged || readFile(t, root, stored.File) != before {
			t.Fatalf("%s changed:\n%s", id, readFile(t, root, stored.File))
		}
	}
}

func TestRecordKeepsUnknownFieldsAndComments(t *testing.T) {
	project, root := newProject(t)
	original := readFile(t, root, capabilityFile)
	writeFile(t, root, capabilityFile, strings.Replace(
		original,
		"serves: bc-0001\n",
		"serves: bc-0001 # the payments goal\nextra: kept\n",
		1,
	))
	fields := storedFields(t, project, "bc-7f3a")
	fields["intent"] = "Sign in with email and password."

	save(t, project, payloadOf(t, "capability", fields))

	text := readFile(t, root, capabilityFile)
	for _, want := range []string{
		"intent: Sign in with email and password.\n",
		"serves: bc-0001 # the payments goal\n",
		"extra: kept\n",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
}

func TestRecordFailsLikeTheEquivalentFile(t *testing.T) {
	project, _ := newProject(t)
	fields := storedFields(t, project, "bc-e5f6")
	delete(fields, "status")

	_, err := project.Prepare(
		context.Background(),
		payloadOf(t, "question", fields),
		author,
	)

	view := requireCode(t, err, string(spider.SavePayloadInvalid))
	findings, _ := view.Details["findings"].([]spider.Finding)
	if len(findings) != 1 || findings[0].Code != "question_field_missing" {
		t.Fatalf("findings = %+v", findings)
	}
}

func TestRecordCreatesWithBinaryOwnedFields(t *testing.T) {
	project, root := newProject(t)

	prepared := save(t, project, payloadOf(t, "decision", map[string]any{
		"what":    "Store sessions in Redis",
		"why":     "Share state across instances.",
		"status":  "active",
		"created": "2020-01-01",
	}))

	if !strings.HasPrefix(prepared.File, ".beacon/records/decisions/") ||
		!strings.HasSuffix(prepared.File, "-store-sessions-in-redis.yaml") {
		t.Fatalf("file = %s", prepared.File)
	}
	if len(prepared.Ignored) != 1 || prepared.Ignored[0] != "created" {
		t.Fatalf("ignored = %v", prepared.Ignored)
	}
	text := readFile(t, root, prepared.File)
	if !strings.Contains(text, "author: \""+author+"\"") ||
		strings.Contains(text, "2020-01-01") {
		t.Fatalf("record:\n%s", text)
	}
}

func TestRecordIssuesIDsForNewPromisesAndTasks(t *testing.T) {
	project, _ := newProject(t)
	fields := storedFields(t, project, "bc-7f3a")
	fields["promises"] = append(promisesOf(fields), map[string]any{
		"text":   "An unknown email gets the same response as a valid one.",
		"check":  map[string]any{"manual": "Try a made-up email"},
		"status": "proposed",
	})

	save(t, project, payloadOf(t, "capability", fields))
	task := save(t, project, payloadOf(t, "task", map[string]any{
		"title":  "Migrate the data",
		"parent": "bc-91cd",
		"status": "open",
	}))

	promises := promisesOf(storedFields(t, project, "bc-7f3a"))
	if len(promises) != 3 {
		t.Fatalf("promises = %v", promises)
	}
	if task.ID != "bc-91cd.2" {
		t.Fatalf("task id = %s", task.ID)
	}
}

func TestRecordRejectsOmittedAndUnknownPromises(t *testing.T) {
	project, _ := newProject(t)
	fields := storedFields(t, project, "bc-7f3a")
	promises := promisesOf(fields)

	fields["promises"] = promises[:1]
	_, err := project.Prepare(
		context.Background(),
		payloadOf(t, "capability", fields),
		author,
	)
	requireCode(t, err, "promise_omitted")

	invented := map[string]any{"id": "bc-eeee", "text": "x", "status": "active"}
	fields["promises"] = append(promises, invented)
	_, err = project.Prepare(
		context.Background(),
		payloadOf(t, "capability", fields),
		author,
	)
	requireCode(t, err, "promise_unknown")
}

func TestRecordRejectsABrokenReference(t *testing.T) {
	project, _ := newProject(t)
	fields := storedFields(t, project, "bc-7f3a")
	fields["serves"] = "bc-9999"

	_, err := project.Prepare(
		context.Background(),
		payloadOf(t, "capability", fields),
		author,
	)

	requireCode(t, err, string(spider.SaveGraphInvalid))
}

func TestRecordRejectsAMissingID(t *testing.T) {
	project, _ := newProject(t)

	_, err := project.Prepare(
		context.Background(),
		payloadOf(t, "goal", map[string]any{"id": "bc-ffff"}),
		author,
	)

	requireCode(t, err, "record_missing")
}

func TestSaveStampsTheApprovedPromise(t *testing.T) {
	project, root := newProject(t)
	fields := storedFields(t, project, "bc-7f3a")
	promise := promisesOf(fields)[0].(map[string]any)
	promise["text"] = "After 10 failed attempts the account locks."
	prepared, err := project.Prepare(
		context.Background(),
		payloadOf(t, "capability", fields),
		author,
	)
	if err != nil {
		t.Fatal(err)
	}

	stamps := map[string]spider.Approval{
		"bc-c3d4": {By: "human:jane", At: "2026-11-25", Via: author},
	}
	err = project.Save(context.Background(), prepared, stamps)
	if err != nil {
		t.Fatal(err)
	}

	want := "approved: {by: \"human:jane\", at: 2026-11-25, via: \"" +
		author + "\"}"
	text := readFile(t, root, capabilityFile)
	if !strings.Contains(text, want) {
		t.Fatalf("missing %q in:\n%s", want, text)
	}
}

func TestSaveRefusesARecordChangedSinceRead(t *testing.T) {
	project, root := newProject(t)
	fields := storedFields(t, project, "bc-7f3a")
	fields["name"] = "Login"
	prepared, err := project.Prepare(
		context.Background(),
		payloadOf(t, "capability", fields),
		author,
	)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, capabilityFile, readFile(t, root, capabilityFile)+"\n")

	err = project.Save(context.Background(), prepared, nil)

	requireCode(t, err, "record_changed")
}

func TestRecordUpdatesAGuideBody(t *testing.T) {
	project, root := newProject(t)
	fields := storedFields(t, project, "bc-0a9b")
	fields["body"] = "Un solo servicio.\n"

	prepared := save(t, project, payloadOf(t, "guide", fields))

	text := readFile(t, root, prepared.File)
	if !strings.HasSuffix(text, "---\nUn solo servicio.\n") {
		t.Fatalf("guide:\n%s", text)
	}
}

func TestRecordCreatesASubtaskFromItsID(t *testing.T) {
	project, _ := newProject(t)

	task := save(t, project, payloadOf(t, "task", map[string]any{
		"id":     "bc-91cd.1.2",
		"title":  "Create the indexes",
		"parent": "bc-91cd",
		"status": "open",
	}))

	if task.ID != "bc-91cd.1.2" || task.Before != nil {
		t.Fatalf("task = %+v", task)
	}
}
