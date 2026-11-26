package facade

import (
	"encoding/json"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

func recordOf(kind, body string) *spider.Record {
	return &spider.Record{Type: kind, Data: json.RawMessage(body)}
}

const lockoutBody = `{"id":"bc-7f3a","name":"Login","promises":[
{"id":"bc-c3d4","text":"Locks after 5","status":"active",
"check":{"runner":"test-go","target":"auth"}}]}`

func TestConvertChangeClassifiesAnEditedActivePromise(t *testing.T) {
	edited := `{"id":"bc-7f3a","name":"Login","promises":[
{"id":"bc-c3d4","text":"Locks after 10","status":"active",
"check":{"runner":"test-go","target":"auth"}}]}`

	change, err := convertChange(spider.StateChange{
		Path:   ".beacon/records/capabilities/bc-7f3a-login.yaml",
		Kind:   "modified",
		Before: recordOf("capability", lockoutBody),
		After:  recordOf("capability", edited),
	})

	if err != nil {
		t.Fatal(err)
	}

	level, needsYes := deltas.ClassifyChange(change)
	if level != deltas.LevelWeakening || !needsYes {
		t.Fatalf("got %d/%v", level, needsYes)
	}
}

func TestConvertChangeKeepsTheApprovalStamp(t *testing.T) {
	stamped := `{"id":"bc-0001","status":"dropped",
"approved":{"by":"human:jane","at":"2026-11-26"}}`
	plain := `{"id":"bc-0001","status":"active"}`

	change, err := convertChange(spider.StateChange{
		Path:   "goal.yaml",
		Kind:   "modified",
		Before: recordOf("goal", plain),
		After:  recordOf("goal", stamped),
	})
	if err != nil {
		t.Fatal(err)
	}

	id, err := kernel.ParseRecordID("bc-0001")
	if err != nil {
		t.Fatal(err)
	}

	target := deltas.RecordTarget(id)

	if !deltas.IsFreshlyApproved(target, []deltas.StateChange{change}) {
		t.Fatal("the stamp was lost in conversion")
	}
}

func TestConvertChangeRejectsAnUnknownRecordType(t *testing.T) {
	_, err := convertChange(spider.StateChange{
		Path:  "x.yaml",
		Kind:  "added",
		After: recordOf("sketch", `{"id":"bc-0001"}`),
	})

	if err == nil {
		t.Fatal("expected an error")
	}
}
