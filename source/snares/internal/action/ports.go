package action

import (
	"context"

	"github.com/ronz204/beacon/source/spider"
	"github.com/ronz204/beacon/source/weaver"
)

type Ports interface {
	Doctor(ctx context.Context) []spider.Finding
	HasHead(ctx context.Context) bool
	Diff(
		ctx context.Context,
		since, base string,
	) (weaver.DiffView, []spider.Signal, error)
	Check(
		ctx context.Context,
		capability string,
		all bool,
		base string,
	) (weaver.CheckView, []spider.Signal, error)
}

type CIPorts interface {
	Ports
	CheckAt(
		ctx context.Context,
		all bool,
		base string,
		runnersAt string,
	) (weaver.CheckView, []spider.Signal, error)
}

type Block struct {
	Kind    string
	Message string
	Hint    string
	Items   []string
}

const (
	BlockInvalidState = "invalid_state"
	BlockUnapproved   = "unapproved_change"
	BlockFailedCheck  = "failed_check"
	BlockRunners      = "runners_unapproved"
)

const runnersChanged = "runners_changed"

const (
	invalidStateHint = "fix the reported records, then run beacon doctor"
	unapprovedHint   = "ask the person to approve the level 3 changes; if " +
		"they say yes, send the record again with beacon record --approve"
	runnersHint = "the pull request changes the runners without a recorded " +
		"yes: the promises that depend on them did not run; ask the person " +
		"to approve the config change"
	failedCheckHint = "fix the code until beacon check passes; do not edit " +
		"the promise"
)
