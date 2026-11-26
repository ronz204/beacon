package anchor

import (
	"errors"
	"testing"
)

func TestReviewWarnsWhenTheBinaryIsNotOnPath(t *testing.T) {
	missing := func(string) (string, error) {
		return "", errors.New("not found")
	}

	notes := reviewNotes(missing)
	if len(notes) != 1 || notes[0] != missingBinaryNote {
		t.Errorf("notes = %v", notes)
	}
}

func TestReviewIsEmptyWhenTheBinaryIsOnPath(t *testing.T) {
	found := func(name string) (string, error) {
		return name, nil
	}

	if notes := reviewNotes(found); len(notes) != 0 {
		t.Errorf("notes = %v", notes)
	}
}
