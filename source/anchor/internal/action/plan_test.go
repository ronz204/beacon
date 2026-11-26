package action_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/assets"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
	"github.com/ronz204/beacon/source/anchor/internal/driver/double"
)

type sampleAgent struct {
	covers []agents.Moment
}

func (s sampleAgent) Name() string {
	return "sample-agent"
}

func (s sampleAgent) Moments() []agents.Moment {
	return s.covers
}

func (s sampleAgent) Contribute(
	agents.Options,
) (agents.Contribution, error) {
	importBlock, err := blocks.ParseBlock(
		"<!-- BEGIN BEACON -->\n@AGENTS.md\n<!-- END BEACON -->",
	)
	if err != nil {
		return agents.Contribution{}, err
	}

	allow, err := blocks.ParseEntry(
		blocks.EntryAppend, "Bash(beacon *)", "permissions", "allow",
	)
	if err != nil {
		return agents.Contribution{}, err
	}

	return agents.Contribution{
		Blocks: []agents.BlockFile{{
			Path:  "CLAUDE.md",
			Part:  agents.PartImport,
			Block: importBlock,
		}},
		Entries: []agents.EntryFile{{
			Path:    ".claude/settings.json",
			Part:    agents.PartPermission,
			Entries: []blocks.Entry{allow},
		}},
	}, nil
}

func newPlanner(
	t *testing.T,
	files *double.Files,
	list ...agents.Agent,
) action.Planner {
	t.Helper()

	method, err := blocks.ParseBlock(assets.Method)
	if err != nil {
		t.Fatal(err)
	}

	shared := agents.Contribution{
		Blocks: []agents.BlockFile{{
			Path:  "AGENTS.md",
			Part:  agents.PartGuide,
			Block: method,
		}},
	}

	return action.Planner{Files: files, Shared: shared, Agents: list}
}

func TestPlanOfAnEmptyRepoCreatesTheGuide(t *testing.T) {
	files := double.NewFiles(map[string]string{})

	planner := newPlanner(t, files)

	plan, err := planner.Plan(context.Background(), agents.Options{})
	if err != nil {
		t.Fatal(err)
	}

	if len(plan.Changes) != 1 {
		t.Fatalf("changes = %d, want 1", len(plan.Changes))
	}

	change := plan.Changes[0]
	if change.Path != "AGENTS.md" || !change.Created {
		t.Errorf("change = %+v", change)
	}
	if !slices.Equal(change.Parts, []agents.Part{agents.PartGuide}) {
		t.Errorf("parts = %v", change.Parts)
	}
	if len(files.Writes) != 0 {
		t.Errorf("planning wrote %v", files.Writes)
	}
}

func TestApplyThenPlanAgainHasNothingToDo(t *testing.T) {
	ctx := context.Background()
	files := double.NewFiles(map[string]string{"CLAUDE.md": "# Mine\n"})
	planner := newPlanner(t, files, sampleAgent{})

	plan, err := planner.Plan(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}

	applier := action.Applier{Files: files, Stamp: "stamp"}
	if err := applier.Apply(ctx, plan); err != nil {
		t.Fatal(err)
	}

	again, err := planner.Plan(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}

	if len(again.Changes) != 0 {
		t.Errorf("second plan has %d changes", len(again.Changes))
	}
}

func TestApplyBacksUpExistingFilesOnly(t *testing.T) {
	ctx := context.Background()
	files := double.NewFiles(map[string]string{"CLAUDE.md": "# Mine\n"})
	planner := newPlanner(t, files, sampleAgent{})

	plan, err := planner.Plan(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}

	applier := action.Applier{Files: files, Stamp: "stamp"}
	if err := applier.Apply(ctx, plan); err != nil {
		t.Fatal(err)
	}

	if files.Backups["stamp/CLAUDE.md"] != "# Mine\n" {
		t.Errorf("backups = %v", files.Backups)
	}
	if len(files.Backups) != 1 {
		t.Errorf("backed up %d files, want 1", len(files.Backups))
	}
}

func TestPlanNamesTheMomentsAnAgentCannotCover(t *testing.T) {
	files := double.NewFiles(map[string]string{})
	covers := []agents.Moment{agents.MomentStart, agents.MomentStop}

	plan, err := newPlanner(t, files, sampleAgent{covers: covers}).Plan(
		context.Background(),
		agents.Options{},
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []agents.Moment{agents.MomentCompact, agents.MomentEdit}
	if !slices.Equal(plan.Agents[0].Missing, want) {
		t.Errorf("missing = %v, want %v", plan.Agents[0].Missing, want)
	}
}

func TestPlanFailsOnAnInvalidSettingsFile(t *testing.T) {
	files := double.NewFiles(map[string]string{
		".claude/settings.json": "{nope",
	})

	_, err := newPlanner(t, files, sampleAgent{}).Plan(
		context.Background(),
		agents.Options{},
	)

	var failure blocks.EntryError
	if !errors.As(err, &failure) || failure.Code != blocks.DocumentInvalid {
		t.Errorf("err = %v", err)
	}
}

func TestApplyWritesNothingWhenABackupFails(t *testing.T) {
	ctx := context.Background()
	files := double.NewFiles(map[string]string{"AGENTS.md": "# Mine\n"})

	plan, err := newPlanner(t, files).Plan(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}

	files.Err = errors.New("disk full")
	err = action.Applier{Files: files, Stamp: "stamp"}.Apply(ctx, plan)

	if err == nil || len(files.Writes) != 0 {
		t.Errorf("err = %v, writes = %v", err, files.Writes)
	}
}

func TestWithdrawDeletesFilesThatOnlyHeldWhatWasInstalled(t *testing.T) {
	ctx := context.Background()
	files := double.NewFiles(map[string]string{})
	planner := newPlanner(t, files, sampleAgent{})

	plan, err := planner.Plan(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}

	applier := action.Applier{Files: files, Stamp: "stamp"}
	if err := applier.Apply(ctx, plan); err != nil {
		t.Fatal(err)
	}

	removal, err := planner.Withdraw(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := applier.Apply(ctx, removal); err != nil {
		t.Fatal(err)
	}

	if len(files.Texts) != 0 {
		t.Errorf("files left behind: %v", files.Texts)
	}
}

func TestWithdrawKeepsASettingsFileThatHasOtherContent(t *testing.T) {
	ctx := context.Background()
	original := "{\n  \"model\": \"sample\"\n}\n"
	files := double.NewFiles(map[string]string{
		".claude/settings.json": original,
	})
	planner := newPlanner(t, files, sampleAgent{})
	applier := action.Applier{Files: files, Stamp: "stamp"}

	plan, err := planner.Plan(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := applier.Apply(ctx, plan); err != nil {
		t.Fatal(err)
	}

	removal, err := planner.Withdraw(ctx, agents.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := applier.Apply(ctx, removal); err != nil {
		t.Fatal(err)
	}

	if files.Texts[".claude/settings.json"] != original {
		t.Errorf("settings = %q", files.Texts[".claude/settings.json"])
	}
}
