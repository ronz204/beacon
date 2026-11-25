package record_test

import (
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

func recordName(t *testing.T, id, title, extension string) string {
	t.Helper()

	parsedID, err := kernel.ParseRecordID(id)
	if err != nil {
		t.Fatal(err)
	}

	parsedTitle, err := kernel.ParseRecordText(title)
	if err != nil {
		t.Fatal(err)
	}

	return record.FileName(parsedID, parsedTitle, extension)
}

func TestFileNameBuildsReadableSlug(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"Café crème brûlée", "bc-a1b2-cafe-creme-brulee.yaml"},
		{"Naïve façade?", "bc-a1b2-naive-facade.yaml"},
		{"Piñata & Co.", "bc-a1b2-pinata-co.yaml"},
		{"Über Straße", "bc-a1b2-uber-strasse.yaml"},
		{"  JWT   corto v2  ", "bc-a1b2-jwt-corto-v2.yaml"},
		{"src/auth: login", "bc-a1b2-src-auth-login.yaml"},
		{"日本語", "bc-a1b2-record.yaml"},
		{"—?", "bc-a1b2-record.yaml"},
	}

	for _, c := range cases {
		t.Run(c.title, func(t *testing.T) {
			got := recordName(t, "bc-a1b2", c.title, ".yaml")

			if got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestFileNameKeepsChildIDsAndExtension(t *testing.T) {
	got := recordName(t, "bc-91cd.1.1", "Create tables", ".md")

	if got != "bc-91cd.1.1-create-tables.md" {
		t.Fatalf("got %q", got)
	}
}

func TestFileNameLimitsLongTitles(t *testing.T) {
	title := strings.Repeat("palabra ", 30)

	got := recordName(t, "bc-a1b2", title, ".yaml")
	slug := strings.TrimSuffix(strings.TrimPrefix(got, "bc-a1b2-"), ".yaml")

	if len(slug) > 50 {
		t.Fatalf("slug has %d characters", len(slug))
	}
	if strings.HasSuffix(slug, "-") {
		t.Fatalf("slug ends with a hyphen: %q", slug)
	}
	if !strings.HasSuffix(slug, "palabra") {
		t.Fatalf("slug cuts a word: %q", slug)
	}
}

func TestFileNameCutsASingleLongWord(t *testing.T) {
	title := strings.Repeat("a", 80)

	got := recordName(t, "bc-a1b2", title, ".yaml")

	if got != "bc-a1b2-"+strings.Repeat("a", 50)+".yaml" {
		t.Fatalf("got %q", got)
	}
}
