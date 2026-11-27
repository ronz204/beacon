package blocks_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

const sampleRule = "---\npaths:\n  - \".beacon/**\"\n---\n\n# Sample\n\n- one\n"

func stamped(t *testing.T) string {
	t.Helper()

	text, err := blocks.StampManaged(sampleRule)
	if err != nil {
		t.Fatal(err)
	}

	return text
}

func TestStampPutsTheDigestRightAfterTheOpeningFence(t *testing.T) {
	lines := strings.Split(stamped(t), "\n")

	if lines[0] != "---" || !strings.HasPrefix(lines[1], "beacon: ") {
		t.Errorf("lines = %q", lines[:3])
	}
	if lines[2] != "paths:" {
		t.Errorf("the frontmatter moved: %q", lines[:4])
	}
}

func TestStampingTheSameContentTwiceGivesTheSameText(t *testing.T) {
	first := stamped(t)
	second := stamped(t)

	if first != second {
		t.Error("the stamp is not deterministic")
	}
}

func TestAFreshStampedFileIsManaged(t *testing.T) {
	got := blocks.InspectManaged(stamped(t))

	if got != blocks.OwnershipManaged {
		t.Errorf("ownership = %s", got)
	}
}

func TestWindowsLineEndingsDoNotCountAsAnEdit(t *testing.T) {
	text := strings.ReplaceAll(stamped(t), "\n", "\r\n")

	got := blocks.InspectManaged(text)

	if got != blocks.OwnershipManaged {
		t.Errorf("ownership = %s", got)
	}
}

func TestAnEditedBodyOrFrontmatterIsEdited(t *testing.T) {
	original := stamped(t)
	extra := strings.Replace(original, "paths:", "paths: []\nx:", 1)
	cases := map[string]string{
		"body":        original + "- mine\n",
		"frontmatter": extra,
	}

	for name, text := range cases {
		if got := blocks.InspectManaged(text); got != blocks.OwnershipEdited {
			t.Errorf("%s: ownership = %s", name, got)
		}
	}
}

func TestAFileWithoutTheStampIsForeign(t *testing.T) {
	cases := []string{"", "# Mine\n", sampleRule}

	for _, text := range cases {
		if got := blocks.InspectManaged(text); got != blocks.OwnershipForeign {
			t.Errorf("%q: ownership = %s", text, got)
		}
	}
}

func TestStampRefusesContentWithoutFrontmatter(t *testing.T) {
	cases := []string{"# Title\n", "---\nnever closed\n", ""}

	for _, text := range cases {
		_, err := blocks.StampManaged(text)

		var failure blocks.ManagedError
		if !errors.As(err, &failure) {
			t.Errorf("%q: err = %v", text, err)
		}
	}
}
