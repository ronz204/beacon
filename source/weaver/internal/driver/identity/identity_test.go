package identity_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ronz204/beacon/source/system"
	"github.com/ronz204/beacon/source/weaver/internal/driver/identity"
)

func newDirectory(t *testing.T, name string) string {
	t.Helper()

	directory := t.TempDir()
	ctx := context.Background()

	if _, err := system.Git(ctx, directory, "init", "-q"); err != nil {
		t.Skipf("git is not available: %v", err)
	}
	if name != "" {
		_, err := system.Git(ctx, directory, "config", "user.name", name)
		if err != nil {
			t.Fatal(err)
		}
	}

	return directory
}

func failingLookup(context.Context) (string, error) {
	return "", errors.New("state unreadable")
}

func terminal(answer bool) func() bool {
	return func() bool {
		return answer
	}
}

func agentNamed(name string) func(context.Context) (string, error) {
	return func(context.Context) (string, error) {
		return name, nil
	}
}

func TestAPersonAtATerminalIsHuman(t *testing.T) {
	t.Setenv("BEACON_ACTOR", "")
	who := identity.Git{
		Directory:   newDirectory(t, "jane"),
		Interactive: terminal(true),
		ActiveAgent: agentNamed("claude-code"),
	}

	person, err := who.Whoami(context.Background())
	if err != nil || person.Actor.String() != "human:jane" {
		t.Errorf("person %+v, err %v", person, err)
	}
}

func TestWithoutATerminalTheActorIsTheActiveAgent(t *testing.T) {
	t.Setenv("BEACON_ACTOR", "")
	who := identity.Git{
		Directory:   newDirectory(t, "jane"),
		Interactive: terminal(false),
		ActiveAgent: agentNamed("claude-code"),
	}

	person, err := who.Whoami(context.Background())

	if err != nil || person.Actor.String() != "agent:claude-code" {
		t.Errorf("person %+v, err %v", person, err)
	}
	if person.Human != "jane" {
		t.Errorf("human = %q", person.Human)
	}
}

func TestWithoutATerminalOrAKnownAgentTheActorIsUnknown(t *testing.T) {
	t.Setenv("BEACON_ACTOR", "")
	directory := newDirectory(t, "jane")

	cases := map[string]func(context.Context) (string, error){
		"none":    nil,
		"empty":   agentNamed(""),
		"failure": failingLookup,
	}

	for name, lookup := range cases {
		who := identity.Git{
			Directory:   directory,
			Interactive: terminal(false),
			ActiveAgent: lookup,
		}

		person, err := who.Whoami(context.Background())
		if err != nil || person.Actor.String() != "agent:unknown" {
			t.Errorf("%s: person %+v, err %v", name, person, err)
		}
	}
}

func TestTheExplicitActorWinsOverEverything(t *testing.T) {
	t.Setenv("BEACON_ACTOR", "agent:sample-agent")
	who := identity.Git{
		Directory:   newDirectory(t, "jane"),
		Interactive: terminal(true),
	}

	person, err := who.Whoami(context.Background())
	if err != nil || person.Actor.String() != "agent:sample-agent" {
		t.Errorf("person %+v, err %v", person, err)
	}
}

func TestAPersonNeedsTheGitName(t *testing.T) {
	t.Setenv("BEACON_ACTOR", "")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	who := identity.Git{
		Directory:   newDirectory(t, ""),
		Interactive: terminal(true),
	}

	if _, err := who.Whoami(context.Background()); err == nil {
		t.Error("expected identity_missing")
	}
}
