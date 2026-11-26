package action

import (
	"context"
	"encoding/json"

	"github.com/ronz204/beacon/source/kernel"
)

type ClaimCode string

const (
	ClaimTaken          ClaimCode = "claim_taken"
	ClaimNotTask        ClaimCode = "claim_not_task"
	ClaimTaskDone       ClaimCode = "claim_task_done"
	ClaimTaskArchived   ClaimCode = "claim_task_archived"
	ClaimBranchDetached ClaimCode = "claim_branch_detached"
)

type ClaimError = kernel.Error[ClaimCode]

var claimErrors = kernel.ErrorCatalog[ClaimCode]{
	ClaimTaken:          "the task is already claimed elsewhere",
	ClaimNotTask:        "only tasks can be claimed",
	ClaimTaskDone:       "the task is already done",
	ClaimTaskArchived:   "the task was closed and archived",
	ClaimBranchDetached: "a task is claimed from a branch, not a bare commit",
}

type StoredTask struct {
	Type     string
	Data     []byte
	Archived bool
}

type BranchState struct {
	Name     string
	Detached bool
}

type LocalClaim struct {
	Task     string
	Branch   string
	Worktree string
	Day      string
}

type ClaimPorts interface {
	Task(ctx context.Context, id string) (StoredTask, error)
	Branch(ctx context.Context) (BranchState, error)
	Worktree() string
	Claims(ctx context.Context) ([]LocalClaim, error)
	SaveClaim(ctx context.Context, claim LocalClaim) error
	RemoveClaim(ctx context.Context, task string) error
}

type ClaimInput struct {
	Task  string
	Force bool
}

type ClaimHolder struct {
	Source   string
	Branch   string
	Worktree string
	Owner    string
}

type ClaimResult struct {
	Task     string
	Branch   string
	Worktree string
	Owner    string
	Replaced *ClaimHolder
}

type Claimer struct {
	Ports    ClaimPorts
	Store    RecordStore
	Identity Identity
}

func (c Claimer) Claim(
	ctx context.Context,
	input ClaimInput,
) (ClaimResult, error) {
	person, err := c.Identity.Whoami(ctx)
	if err != nil {
		return ClaimResult{}, err
	}

	branch, err := c.Ports.Branch(ctx)
	if err != nil {
		return ClaimResult{}, err
	}
	if branch.Detached {
		return ClaimResult{}, claimErrors.Fail(ClaimBranchDetached, nil)
	}

	fields, err := c.openTask(ctx, input.Task)
	if err != nil {
		return ClaimResult{}, err
	}

	local, err := c.Ports.Claims(ctx)
	if err != nil {
		return ClaimResult{}, err
	}

	previous := localClaimOf(local, input.Task)
	holder := c.holderOf(previous, fields, branch.Name)
	if holder != nil && !input.Force {
		return ClaimResult{}, taken(input.Task, *holder)
	}

	result := ClaimResult{
		Task:     input.Task,
		Branch:   branch.Name,
		Worktree: c.Ports.Worktree(),
		Owner:    person.Actor.String(),
		Replaced: holder,
	}

	return result, c.write(ctx, result, previous, fields)
}

func (c Claimer) openTask(
	ctx context.Context,
	id string,
) (map[string]any, error) {
	stored, err := c.Ports.Task(ctx, id)
	if err != nil {
		return nil, err
	}

	details := map[string]any{"task": id}
	if stored.Archived {
		return nil, claimErrors.Fail(ClaimTaskArchived, details)
	}
	if stored.Type != "task" {
		return nil, claimErrors.Fail(ClaimNotTask, details)
	}

	var fields map[string]any
	if err := json.Unmarshal(stored.Data, &fields); err != nil {
		return nil, err
	}
	if fields["status"] == "done" {
		return nil, claimErrors.Fail(ClaimTaskDone, details)
	}

	return fields, nil
}

func (c Claimer) holderOf(
	previous *LocalClaim,
	fields map[string]any,
	branch string,
) *ClaimHolder {
	if previous != nil && !c.isHere(*previous, branch) {
		return &ClaimHolder{
			Source:   "local",
			Branch:   previous.Branch,
			Worktree: previous.Worktree,
		}
	}

	recorded, _ := fields["branch"].(string)
	if fields["status"] == "claimed" && recorded != branch {
		owner, _ := fields["owner"].(string)

		return &ClaimHolder{Source: "record", Branch: recorded, Owner: owner}
	}

	return nil
}

func (c Claimer) isHere(claim LocalClaim, branch string) bool {
	if claim.Branch != branch {
		return false
	}

	return claim.Worktree == c.Ports.Worktree()
}

func (c Claimer) write(
	ctx context.Context,
	result ClaimResult,
	previous *LocalClaim,
	fields map[string]any,
) error {
	claim := LocalClaim{
		Task:     result.Task,
		Branch:   result.Branch,
		Worktree: result.Worktree,
	}
	if err := c.Ports.SaveClaim(ctx, claim); err != nil {
		return err
	}

	if err := c.saveTask(ctx, result, fields); err != nil {
		return c.rollback(ctx, result.Task, previous, err)
	}

	return nil
}

func (c Claimer) saveTask(
	ctx context.Context,
	result ClaimResult,
	fields map[string]any,
) error {
	fields["status"] = "claimed"
	fields["owner"] = result.Owner
	fields["branch"] = result.Branch

	payload, err := json.Marshal(map[string]any{
		"type":   "task",
		"record": fields,
	})
	if err != nil {
		return err
	}

	prepared, err := c.Store.Prepare(ctx, payload, result.Owner)
	if err != nil {
		return err
	}
	if prepared.Unchanged {
		return nil
	}

	return c.Store.Save(ctx, prepared, nil)
}

func (c Claimer) rollback(
	ctx context.Context,
	task string,
	previous *LocalClaim,
	cause error,
) error {
	if previous != nil {
		_ = c.Ports.SaveClaim(ctx, *previous)
		return cause
	}

	_ = c.Ports.RemoveClaim(ctx, task)

	return cause
}

func localClaimOf(claims []LocalClaim, task string) *LocalClaim {
	for _, claim := range claims {
		if claim.Task == task {
			return &claim
		}
	}

	return nil
}

func taken(task string, holder ClaimHolder) ClaimError {
	return claimErrors.Fail(ClaimTaken, map[string]any{
		"task":     task,
		"source":   holder.Source,
		"branch":   holder.Branch,
		"worktree": holder.Worktree,
		"owner":    holder.Owner,
	})
}
