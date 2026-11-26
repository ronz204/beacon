package action_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/weaver/internal/action"
)

var errDisk = errors.New("disk full")

type fakePorts struct {
	claims map[string]action.LocalClaim
	record string
	branch string
	tree   string
}

func (f *fakePorts) Task(
	context.Context,
	string,
) (action.StoredTask, error) {
	return action.StoredTask{Type: "task", Data: []byte(f.record)}, nil
}

func (f *fakePorts) Branch(context.Context) (action.BranchState, error) {
	return action.BranchState{Name: f.branch}, nil
}

func (f *fakePorts) Worktree() string {
	return f.tree
}

func (f *fakePorts) Claims(context.Context) ([]action.LocalClaim, error) {
	claims := make([]action.LocalClaim, 0, len(f.claims))
	for _, claim := range f.claims {
		claims = append(claims, claim)
	}

	return claims, nil
}

func (f *fakePorts) SaveClaim(
	_ context.Context,
	claim action.LocalClaim,
) error {
	f.claims[claim.Task] = claim
	return nil
}

func (f *fakePorts) RemoveClaim(_ context.Context, task string) error {
	delete(f.claims, task)
	return nil
}

type failingStore struct{}

func (failingStore) Prepare(
	context.Context,
	[]byte,
	string,
) (action.PreparedRecord, error) {
	return action.PreparedRecord{}, errDisk
}

func (failingStore) Save(
	context.Context,
	action.PreparedRecord,
	map[string]kernel.ApprovalInput,
) error {
	return errDisk
}

type fixedIdentity struct{}

func (fixedIdentity) Whoami(context.Context) (action.Person, error) {
	actor, err := kernel.ParseActor("agent:sample-agent")

	return action.Person{Actor: actor, Human: "jane"}, err
}

const openTask = `{"id":"bc-91cd.1","status":"open"}`

func newClaimer(ports *fakePorts) action.Claimer {
	return action.Claimer{
		Ports:    ports,
		Store:    failingStore{},
		Identity: fixedIdentity{},
	}
}

func TestClaimUndoesTheLocalClaimWhenTheRecordCannotBeWritten(t *testing.T) {
	ports := &fakePorts{
		claims: map[string]action.LocalClaim{},
		record: openTask,
		branch: "feat/sample-branch",
		tree:   "/work/one",
	}

	_, err := newClaimer(ports).Claim(
		context.Background(),
		action.ClaimInput{Task: "bc-91cd.1"},
	)

	if !errors.Is(err, errDisk) {
		t.Fatalf("err = %v", err)
	}
	if len(ports.claims) != 0 {
		t.Fatalf("claims = %v", ports.claims)
	}
}

func TestForcedClaimRestoresThePreviousHolderOnFailure(t *testing.T) {
	previous := action.LocalClaim{
		Task:     "bc-91cd.1",
		Branch:   "feat/other-branch",
		Worktree: "/work/two",
		Day:      "2026-11-25",
	}
	ports := &fakePorts{
		claims: map[string]action.LocalClaim{"bc-91cd.1": previous},
		record: openTask,
		branch: "feat/sample-branch",
		tree:   "/work/one",
	}

	_, err := newClaimer(ports).Claim(
		context.Background(),
		action.ClaimInput{Task: "bc-91cd.1", Force: true},
	)

	if !errors.Is(err, errDisk) {
		t.Fatalf("err = %v", err)
	}
	if ports.claims["bc-91cd.1"] != previous {
		t.Fatalf("claims = %v", ports.claims)
	}
}
