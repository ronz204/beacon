package anchor

import (
	"errors"
	"testing"
)

func TestReviewWarnsWhenTheBinaryIsNotOnPath(t *testing.T) {
	missing := func(string) (string, error) {
		return "", errors.New("not found")
	}

	notes := Anchor{lookup: missing}.reviewNotes(InstallOptions{})
	if len(notes) != 1 || notes[0] != missingBinaryNote {
		t.Errorf("notes = %v", notes)
	}
}

func TestReviewIsEmptyWhenTheBinaryIsOnPath(t *testing.T) {
	found := func(name string) (string, error) {
		return name, nil
	}

	notes := Anchor{lookup: found}.reviewNotes(InstallOptions{})
	if len(notes) != 0 {
		t.Errorf("notes = %v", notes)
	}
}
