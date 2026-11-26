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

const actorVariable = "BEACON_ACTOR"

type Git struct {
	Directory string
}

func (g Git) Whoami(ctx context.Context) (action.Person, error) {
	name := g.userName(ctx)

	if raw := strings.TrimSpace(os.Getenv(actorVariable)); raw != "" {
		actor, err := kernel.ParseActor(raw)

		return action.Person{Actor: actor, Human: name}, err
	}

	actor, err := kernel.ParseActor("human:" + name)
	if err != nil {
		return action.Person{}, deltas.FailIdentityMissing()
	}

	return action.Person{Actor: actor, Human: name}, nil
}

func (g Git) userName(ctx context.Context) string {
	output, err := system.Git(ctx, g.Directory, "config", "user.name")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(output)
}
