package agenda_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/agenda"
)

var sampleNow = time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func requireCode[C ~string](t *testing.T, err error, want C) {
	t.Helper()

	var got kernel.Error[C]
	if !errors.As(err, &got) {
		t.Fatalf("expected domain error %q, got %v", want, err)
	}
	if got.Code != want {
		t.Fatalf("code = %q, want %q", got.Code, want)
	}
}

func sampleWorkInput() agenda.WorkInput {
	return agenda.WorkInput{
		ID:      must(agenda.ParseWorkID("bc-91cd")),
		Author:  must(kernel.ParseActor("agent:sample-agent")),
		Request: must(agenda.ParseWorkRequest("Move to a new store\n")),
		Implements: []agenda.WorkPromiseID{
			must(agenda.ParseWorkPromiseID("bc-c3d4")),
		},
		Serves: must(agenda.ParseWorkGoalID("bc-0001")),
	}
}

func TestWorkReferenceIDs(t *testing.T) {
	_, err := agenda.ParseWorkID("bc-91cd.1")
	requireCode(t, err, agenda.WorkIDInvalid)

	_, err = agenda.ParseWorkPromiseID("bc-c3d4.1")
	requireCode(t, err, agenda.WorkPromiseInvalid)

	_, err = agenda.ParseWorkGoalID("bc-0001.1")
	requireCode(t, err, agenda.WorkGoalInvalid)
}

func TestParseWorkRequestKeepsTextVerbatim(t *testing.T) {
	const raw = "  Move to a new store\n\n  keep the old one\n"

	got, err := agenda.ParseWorkRequest(raw)

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != raw {
		t.Fatalf("got %q", got)
	}
}

func TestParseWorkRequestRejectsBlank(t *testing.T) {
	blank := []string{
		"",
		"   ",
		"\n\t",
	}

	for _, raw := range blank {
		t.Run(raw, func(t *testing.T) {
			_, err := agenda.ParseWorkRequest(raw)

			requireCode(t, err, agenda.WorkRequestInvalid)
		})
	}
}

func TestParseWorkStatus(t *testing.T) {
	for _, status := range agenda.WorkStatuses() {
		got, err := agenda.ParseWorkStatus(string(status))

		if err != nil {
			t.Fatal(err)
		}
		if got != status {
			t.Fatalf("got %q", got)
		}
	}

	_, err := agenda.ParseWorkStatus("paused")
	requireCode(t, err, agenda.WorkStatusInvalid)
}

func TestNewWorkStartsActiveAndDated(t *testing.T) {
	work, err := agenda.NewWork(sampleWorkInput(), sampleNow)

	if err != nil {
		t.Fatal(err)
	}
	if work.Status() != agenda.WorkActive {
		t.Fatalf("status = %q", work.Status())
	}
	if work.Created().String() != "2026-11-22" {
		t.Fatalf("created = %q", work.Created())
	}
	if len(work.Log()) != 0 {
		t.Fatalf("log = %v", work.Log())
	}
}

func TestNewWorkRejectsMissingField(t *testing.T) {
	input := sampleWorkInput()
	input.Request = agenda.WorkRequest{}

	_, err := agenda.NewWork(input, sampleNow)

	requireCode(t, err, agenda.WorkFieldMissing)
}

func TestWorkProtectsItsLists(t *testing.T) {
	input := sampleWorkInput()
	work := must(agenda.NewWork(input, sampleNow))
	replacement := must(agenda.ParseWorkPromiseID("bc-ffff"))

	input.Implements[0] = replacement
	work.Implements()[0] = replacement

	if work.Implements()[0].String() != "bc-c3d4" {
		t.Fatalf("implements = %v", work.Implements())
	}

	entry := must(kernel.ParseRecordText("Split the schema"))
	log := []kernel.RecordText{entry}
	restored := agenda.RestoreWork(agenda.WorkState{Log: log})
	log[0] = must(kernel.ParseRecordText("changed"))

	if restored.Log()[0] != entry {
		t.Fatalf("log = %v", restored.Log())
	}
}

func TestWorkIdentityAndOptionalFields(t *testing.T) {
	work := must(agenda.NewWork(sampleWorkInput(), sampleNow))
	restored := agenda.RestoreWork(agenda.WorkState{
		ID:     work.ID(),
		Status: agenda.WorkDone,
	})

	if !restored.SameAs(work) {
		t.Fatal("works with the same id should be the same")
	}
	if restored.Status() != agenda.WorkDone {
		t.Fatalf("status = %q", restored.Status())
	}
	if _, found := work.Serves(); !found {
		t.Fatal("serves should be present")
	}
	if _, found := restored.Serves(); found {
		t.Fatal("serves should be absent")
	}
	if _, found := work.Approved(); found {
		t.Fatal("approved should be absent")
	}
}
