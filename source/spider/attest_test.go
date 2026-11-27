package spider_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider"
)

const (
	manualPromise    = "bc-c4a1"
	automaticPromise = "bc-c3d4"
)

func headOf(t *testing.T, root string) string {
	t.Helper()

	command := exec.Command("git", "rev-parse", "--short", "HEAD")
	command.Dir = root

	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(output))
}

func editLogin(t *testing.T, root string) {
	t.Helper()

	path := filepath.Join(root, "src", "auth", "login.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("package auth\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func personStamp() spider.Approval {
	return spider.Approval{By: "human:sam", At: "2026-11-27"}
}

func manualIDs(promises []spider.ManualPromise) []string {
	ids := []string{}
	for _, promise := range promises {
		ids = append(ids, promise.ID)
	}

	return ids
}

func TestManualPromisesListsOnlyWhatChanged(t *testing.T) {
	project, root := newProject(t)
	ctx := context.Background()

	promises, _, err := project.ManualPromises(ctx, headOf(t, root))
	if err != nil || len(promises) != 0 {
		t.Fatalf("clean tree: %v, err %v", manualIDs(promises), err)
	}

	editLogin(t, root)

	promises, _, err = project.ManualPromises(ctx, headOf(t, root))
	if err != nil {
		t.Fatal(err)
	}
	if len(promises) != 1 || promises[0].ID != manualPromise {
		t.Fatalf("after edit: %+v", promises)
	}
	if !promises[0].Drifted || promises[0].Manual == "" {
		t.Errorf("promise = %+v", promises[0])
	}
}

func TestAttestWritesWhoSaidYesAndWhoCarriedIt(t *testing.T) {
	project, root := newProject(t)
	stamp := personStamp()
	stamp.Via = "agent:sample-agent"

	err := project.Attest(
		context.Background(),
		manualPromise,
		stamp,
		headOf(t, root),
		"docs/evidence/login.png",
	)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(root, capabilityFile))
	if err != nil {
		t.Fatal(err)
	}

	text := string(data)
	for _, want := range []string{
		`by: "human:sam"`,
		`via: "agent:sample-agent"`,
		"2026-11-27",
		"docs/evidence/login.png",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
}

func TestAttestedPromiseDriftsWhenItsCodeChanges(t *testing.T) {
	project, root := newProject(t)
	ctx := context.Background()

	err := project.Attest(
		ctx, manualPromise, personStamp(), headOf(t, root), "",
	)
	if err != nil {
		t.Fatal(err)
	}

	promises, _, err := project.ManualPromises(ctx, headOf(t, root))
	if err != nil || len(promises) != 0 {
		t.Fatalf("attested at HEAD: %v, err %v", manualIDs(promises), err)
	}

	editLogin(t, root)

	promises, _, err = project.ManualPromises(ctx, headOf(t, root))
	if err != nil || len(promises) != 1 {
		t.Fatalf("after the edit: %v, err %v", manualIDs(promises), err)
	}
}

func TestAttestRefusesAutomaticAndUnknownPromises(t *testing.T) {
	project, root := newProject(t)
	ctx := context.Background()
	head := headOf(t, root)

	err := project.Attest(ctx, automaticPromise, personStamp(), head, "")

	var failure spider.AttestError
	if !errors.As(err, &failure) || failure.Code != spider.AttestNotManual {
		t.Errorf("automatic promise: err = %v", err)
	}

	err = project.Attest(ctx, "bc-ffff", personStamp(), head, "")
	if kernel.DescribeFailure(err).Code != "record_missing" {
		t.Errorf("unknown promise: err = %v", err)
	}
}

func TestAttestRefusesAnAgentAsTheAttester(t *testing.T) {
	project, root := newProject(t)
	stamp := spider.Approval{By: "agent:sample-agent", At: "2026-11-27"}

	err := project.Attest(
		context.Background(), manualPromise, stamp, headOf(t, root), "",
	)

	if err == nil {
		t.Error("an attestation must come from a person")
	}
}

func TestAttestRefusesCodeThatIsNotCommitted(t *testing.T) {
	project, root := newProject(t)
	editLogin(t, root)

	err := project.Attest(
		context.Background(), manualPromise, personStamp(),
		headOf(t, root), "",
	)

	var failure spider.AttestError
	if !errors.As(err, &failure) || failure.Code != spider.AttestUncommitted {
		t.Errorf("err = %v", err)
	}
}

func TestTouchedCapabilitiesAcceptsAnyRevisionName(t *testing.T) {
	project, root := newProject(t)
	editLogin(t, root)

	touched, err := project.TouchedCapabilities(context.Background(), "HEAD")
	if err != nil || len(touched) != 1 || touched[0] != "bc-7f3a" {
		t.Fatalf("touched %v, err %v", touched, err)
	}

	_, err = project.TouchedCapabilities(context.Background(), "--output=x")
	if err == nil {
		t.Error("a revision that looks like an option must be refused")
	}

	_, err = project.TouchedCapabilities(context.Background(), "no-such-rev")
	if err == nil {
		t.Error("an unknown revision must fail")
	}
}
