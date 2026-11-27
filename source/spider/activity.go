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

type InjectionCode string

const (
	InjectionInvalid InjectionCode = "injection_invalid"
	AgentInvalid     InjectionCode = "agent_invalid"
)

type InjectionError = kernel.Error[InjectionCode]

var injectionErrors = kernel.ErrorCatalog[InjectionCode]{
	InjectionInvalid: "an injection needs a lowercase event name " +
		"and zero or more lines",
	AgentInvalid: "an agent name needs lowercase letters, " +
		"digits and hyphens",
}

var injectionEvent = regexp.MustCompile(`^[a-z][a-z-]{0,31}$`)

func (p Project) RecordInjection(
	ctx context.Context,
	session, event string,
	lines int,
) error {
	id, err := kernel.ParseSessionID(session)
	if err != nil {
		return err
	}
	if !injectionEvent.MatchString(event) || lines < 0 {
		details := map[string]any{"event": event, "lines": lines}
		return injectionErrors.Fail(InjectionInvalid, details)
	}

	log, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	return log.RecordInjection(ctx, id, event, lines)
}
