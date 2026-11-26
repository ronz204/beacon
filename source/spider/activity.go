package spider

import (
	"context"
	"regexp"
	"time"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

var agentName = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

func (p Project) NoteAgent(
	ctx context.Context,
	agent, session string,
) error {
	if !agentName.MatchString(agent) {
		details := map[string]any{"agent": agent}
		return injectionErrors.Fail(AgentInvalid, details)
	}
	if _, err := kernel.ParseSessionID(session); err != nil {
		return err
	}

	local, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	return local.NoteAgent(ctx, p.root, agent, session, time.Now())
}

func (p Project) ActiveAgent(ctx context.Context) (string, error) {
	local, err := states.Open(ctx, p.root)
	if err != nil {
		return "", err
	}

	return local.ActiveAgent(ctx, p.root, time.Now())
}
