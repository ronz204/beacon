package random_test

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"testing"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/driver/random"
)

func requireIssueFailure(t *testing.T, err error) {
	t.Helper()

	var got kernel.KernelError
	if !errors.As(err, &got) {
		t.Fatalf("expected KernelError, got %v", err)
	}
	if got.Code != kernel.KernelIDIssueFailed {
		t.Fatalf("code = %q", got.Code)
	}
}

func free(kernel.RecordID) bool {
	return false
}

func TestIssuerCreatesRootIDWithFourHexCharacters(t *testing.T) {
	pattern := regexp.MustCompile(`^bc-[0-9a-f]{4}$`)

	id, err := random.NewIssuer().IssueRecordID(free)

	if err != nil {
		t.Fatal(err)
	}
	if !pattern.MatchString(id.String()) {
		t.Fatalf("id = %q", id)
	}
}

func TestIssuerRetriesOnCollision(t *testing.T) {
	source := bytes.NewReader([]byte{0xa1, 0xb2, 0xc3, 0xd4})
	issuer := random.NewIssuerFrom(source)
	taken := func(id kernel.RecordID) bool {
		return id.String() == "bc-a1b2"
	}

	id, err := issuer.IssueRecordID(taken)

	if err != nil {
		t.Fatal(err)
	}
	if id.String() != "bc-c3d4" {
		t.Fatalf("id = %q", id)
	}
}

func TestIssuerGrowsWhenTheSpaceIsFull(t *testing.T) {
	taken := func(id kernel.RecordID) bool {
		return len(id.String()) == len("bc-a1b2")
	}

	id, err := random.NewIssuer().IssueRecordID(taken)

	if err != nil {
		t.Fatal(err)
	}
	if len(id.String()) != len("bc-a1b2c") {
		t.Fatalf("id = %q", id)
	}
}

func TestIssuerFailsWhenNothingIsFree(t *testing.T) {
	taken := func(kernel.RecordID) bool {
		return true
	}

	_, err := random.NewIssuer().IssueRecordID(taken)

	requireIssueFailure(t, err)
}

func TestIssuerReportsSourceFailure(t *testing.T) {
	issuer := random.NewIssuerFrom(bytes.NewReader(nil))

	_, err := issuer.IssueRecordID(free)

	requireIssueFailure(t, err)
	if !errors.Is(err, io.EOF) {
		t.Fatalf("cause is not reachable: %v", err)
	}
}
