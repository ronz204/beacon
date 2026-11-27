package identity

import (
	"context"
	"os"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/system"
	"github.com/ronz204/beacon/source/weaver/internal/action"
	"github.com/ronz204/beacon/source/weaver/internal/domain/deltas"
)

const (
	actorVariable = "BEACON_ACTOR"
	unknownAgent  = "unknown"
)

type Git struct {
	Directory   string
	Interactive func() bool
	ActiveAgent func(context.Context) (string, error)
}

func (g Git) Whoami(ctx context.Context) (action.Person, error) {
	name := g.userName(ctx)

	if raw := strings.TrimSpace(os.Getenv(actorVariable)); raw != "" {
		actor, err := kernel.ParseActor(raw)

		return action.Person{Actor: actor, Human: name}, err
	}

	if g.interactive() {
		return personOf("human:"+name, name)
	}

	return personOf("agent:"+g.agentName(ctx), name)
}

func personOf(raw, name string) (action.Person, error) {
	actor, err := kernel.ParseActor(raw)
	if err != nil {
		return action.Person{}, deltas.FailIdentityMissing()
	}

	return action.Person{Actor: actor, Human: name}, nil
}

func (g Git) interactive() bool {
	if g.Interactive != nil {
		return g.Interactive()
	}

	return StdoutIsTerminal()
}

func StdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()

	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func (g Git) agentName(ctx context.Context) string {
	if g.ActiveAgent == nil {
		return unknownAgent
	}

	name, err := g.ActiveAgent(ctx)
	if err != nil || name == "" {
		return unknownAgent
	}

	return name
}

func (g Git) userName(ctx context.Context) string {
	output, err := system.Git(ctx, g.Directory, "config", "user.name")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(output)
}
