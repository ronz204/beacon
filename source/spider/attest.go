package spider

import (
	"context"
	"slices"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/action"
	"github.com/ronz204/beacon/source/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/source/spider/internal/driver/commit"
	"github.com/ronz204/beacon/source/spider/internal/driver/match"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
	"github.com/ronz204/beacon/source/spider/internal/driver/states"
)

type AttestCode string

const (
	AttestNotManual   AttestCode = "attest_not_manual"
	AttestUncommitted AttestCode = "attest_uncommitted"
)

type AttestError = kernel.Error[AttestCode]

var attestErrors = kernel.ErrorCatalog[AttestCode]{
	AttestNotManual:   "only a manual promise can be attested",
	AttestUncommitted: "the code of the promise has uncommitted changes",
}

type ManualPromise struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Text       string `json:"text"`
	Manual     string `json:"manual"`
	State      string `json:"state"`
	Drifted    bool   `json:"drifted"`
}

func (p Project) ManualPromises(
	ctx context.Context,
	revision string,
) ([]ManualPromise, []Finding, error) {
	touched, err := p.TouchedCapabilities(ctx, revision)
	if err != nil {
		return nil, nil, err
	}

	targets, findings := p.CheckTargets(ctx)
	analysis, more := p.Analyze(ctx)
	findings = append(findings, more...)

	states := make(map[string]PromiseView, len(analysis.Promises))
	for _, view := range analysis.Promises {
		states[view.ID] = view
	}

	manual := []ManualPromise{}

	for _, target := range targets {
		if !slices.Contains(touched, target.ID) {
			continue
		}

		for _, promise := range target.Promises {
			view := states[promise.ID]
			if promise.Manual == "" || promise.Status != "active" {
				continue
			}
			if holds(view) {
				continue
			}

			manual = append(manual, ManualPromise{
				ID:         promise.ID,
				Capability: target.ID,
				Text:       view.Text,
				Manual:     promise.Manual,
				State:      view.State,
				Drifted:    view.Drifted,
			})
		}
	}

	return manual, findings, nil
}

func holds(view PromiseView) bool {
	attested := view.State == string(ledger.VerificationAttested)

	return attested && !view.Drifted
}

func (p Project) Attest(
	ctx context.Context,
	promise string,
	stamp Approval,
	commit string,
	evidence string,
) error {
	approval, err := kernel.ParseApproval(stamp)
	if err != nil {
		return err
	}

	attestation, err := ledger.ParsePromiseAttestation(
		ledger.PromiseAttestationInput{
			By:       approval.By().String(),
			At:       approval.At().String(),
			Via:      viaOf(approval),
			Commit:   commit,
			Evidence: evidence,
		},
	)
	if err != nil {
		return err
	}

	capability, err := p.manualOwner(ctx, promise)
	if err != nil {
		return err
	}

	if err := p.requireCommitted(ctx, capability, commit); err != nil {
		return err
	}

	owner, err := kernel.ParseRecordID(capability.ID().String())
	if err != nil {
		return err
	}

	file, err := p.store.Lookup(ctx, owner)
	if err != nil {
		return err
	}

	draft, failures := p.store.DraftAttesting(file, promise, attestation)
	if len(failures) > 0 {
		return p.draftError(failures)
	}

	locks, err := states.Open(ctx, p.root)
	if err != nil {
		return err
	}

	return p.store.Write(draft, locks.LockFile(recordsLock))
}

func viaOf(approval kernel.Approval) string {
	via, found := approval.Via()
	if !found {
		return ""
	}

	return via.String()
}

func (p Project) manualOwner(
	ctx context.Context,
	promise string,
) (ledger.Capability, error) {
	snapshot := action.Load(ctx, p.repositories())

	for _, capability := range snapshot.Capabilities {
		for _, candidate := range capability.Promises() {
			if candidate.ID().String() != promise {
				continue
			}
			if candidate.Check().Manual() == "" {
				return ledger.Capability{}, attestErrors.Fail(
					AttestNotManual,
					map[string]any{"id": promise},
				)
			}

			return capability, nil
		}
	}

	return ledger.Capability{}, record.MissingFailure(promise)
}

func (p Project) requireCommitted(
	ctx context.Context,
	capability ledger.Capability,
	commitID string,
) error {
	files, err := commit.NewRepository(p.root).FilesChangedSince(ctx, commitID)
	if err != nil {
		return err
	}

	dirty, err := touches(match.NewMatcher(), capability.Paths(), files)
	if err != nil {
		return err
	}
	if dirty {
		return attestErrors.Fail(
			AttestUncommitted,
			map[string]any{"capability": capability.ID().String()},
		)
	}

	return nil
}
