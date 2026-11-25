package memory_test

import (
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/memory"
)

func sampleGuideState() memory.GuideState {
	return memory.GuideState{
		ID:    must(memory.ParseGuideID("bc-0a9b")),
		Title: sampleText("General architecture"),
		Kind:  memory.GuideArchitecture,
		Applies: []kernel.PathGlob{
			must(kernel.ParsePathGlob("src/**")),
		},
		Always: true,
		Tags: []memory.GuideTag{
			must(memory.ParseGuideTag("backend")),
		},
		Refers: []memory.GuideCapabilityID{
			must(memory.ParseGuideCapabilityID("bc-7f3a")),
		},
	}
}

func TestGuideClosedSets(t *testing.T) {
	for _, kind := range memory.GuideKinds() {
		if got := must(memory.ParseGuideKind(string(kind))); got != kind {
			t.Fatalf("kind = %q", got)
		}
	}
	for _, audience := range memory.GuideAudiences() {
		got := must(memory.ParseGuideAudience(string(audience)))
		if got != audience {
			t.Fatalf("audience = %q", got)
		}
	}
	for _, status := range memory.GuideStatuses() {
		got := must(memory.ParseGuideStatus(string(status)))
		if got != status {
			t.Fatalf("status = %q", got)
		}
	}

	_, err := memory.ParseGuideKind("notes")
	requireCode(t, err, memory.GuideKindInvalid)

	_, err = memory.ParseGuideAudience("everyone")
	requireCode(t, err, memory.GuideAudienceInvalid)

	_, err = memory.ParseGuideStatus("archived")
	requireCode(t, err, memory.GuideStatusInvalid)
}

func TestGuideReferenceIDs(t *testing.T) {
	_, err := memory.ParseGuideID("bc-0a9b.1")
	requireCode(t, err, memory.GuideIDInvalid)

	_, err = memory.ParseGuideCapabilityID("bc-7f3a.1")
	requireCode(t, err, memory.GuideReferInvalid)
}

func TestParseGuideTag(t *testing.T) {
	got, err := memory.ParseGuideTag(" backend ")

	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "backend" {
		t.Fatalf("got %q", got)
	}

	invalid := []string{
		"",
		"   ",
		"two words",
	}

	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			_, err := memory.ParseGuideTag(raw)

			requireCode(t, err, memory.GuideTagInvalid)
		})
	}
}

func TestNewGuideNeedsOnlyTitleAndKind(t *testing.T) {
	guide, err := memory.NewGuide(memory.GuideState{
		ID:    must(memory.ParseGuideID("bc-0a9b")),
		Title: sampleText("General architecture"),
		Kind:  memory.GuideArchitecture,
	})

	if err != nil {
		t.Fatal(err)
	}
	if guide.Always() {
		t.Fatal("always should default to false")
	}
	if _, found := guide.Review(); found {
		t.Fatal("review should be absent")
	}
	if _, found := guide.Reviewed(); found {
		t.Fatal("reviewed should be absent")
	}
	if _, found := guide.Audience(); found {
		t.Fatal("audience should be absent")
	}
	if _, found := guide.Status(); found {
		t.Fatal("status should be absent")
	}
}

func TestNewGuideRejectsMissingField(t *testing.T) {
	missingTitle := sampleGuideState()
	missingTitle.Title = kernel.RecordText{}
	missingKind := sampleGuideState()
	missingKind.Kind = ""

	for _, state := range []memory.GuideState{missingTitle, missingKind} {
		_, err := memory.NewGuide(state)

		requireCode(t, err, memory.GuideFieldMissing)
	}
}

func TestGuideProtectsItsLists(t *testing.T) {
	state := sampleGuideState()
	guide := must(memory.NewGuide(state))
	glob := must(kernel.ParsePathGlob("changed/**"))
	tag := must(memory.ParseGuideTag("changed"))
	refer := must(memory.ParseGuideCapabilityID("bc-ffff"))

	state.Applies[0] = glob
	state.Tags[0] = tag
	state.Refers[0] = refer
	guide.Applies()[0] = glob
	guide.Tags()[0] = tag
	guide.Refers()[0] = refer

	if guide.Applies()[0].String() != "src/**" {
		t.Fatalf("applies = %v", guide.Applies())
	}
	if guide.Tags()[0].String() != "backend" {
		t.Fatalf("tags = %v", guide.Tags())
	}
	if guide.Refers()[0].String() != "bc-7f3a" {
		t.Fatalf("refers = %v", guide.Refers())
	}
}

func TestRestoreGuideKeepsIdentityAndFields(t *testing.T) {
	guide := must(memory.NewGuide(sampleGuideState()))
	restored := memory.RestoreGuide(memory.GuideState{
		ID:       guide.ID(),
		Title:    guide.Title(),
		Kind:     guide.Kind(),
		Review:   must(memory.ParseReviewSpan("12mo")),
		Reviewed: must(kernel.ParseCalendarDate("2026-11-22")),
		Audience: memory.GuideForAgent,
		Status:   memory.GuideDraft,
	})

	if !restored.SameAs(guide) {
		t.Fatal("guides with the same id should be the same")
	}
	if _, found := restored.Review(); !found {
		t.Fatal("review should be kept")
	}
	if _, found := restored.Reviewed(); !found {
		t.Fatal("reviewed should be kept")
	}
	if audience, _ := restored.Audience(); audience != memory.GuideForAgent {
		t.Fatalf("audience = %q", audience)
	}
	if status, _ := restored.Status(); status != memory.GuideDraft {
		t.Fatalf("status = %q", status)
	}
}
