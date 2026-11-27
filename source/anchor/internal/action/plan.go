package action

import (
	"context"
	"maps"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/anchor/internal/agents"
	"github.com/ronz204/beacon/source/anchor/internal/domain/blocks"
)

type Change struct {
	Path    string
	Created bool
	Deleted bool
	Parts   []agents.Part
	Before  string
	After   string
}

type AgentPlan struct {
	Name    string
	Covers  []agents.Moment
	Missing []agents.Moment
}

type Kept struct {
	Path string
	Part agents.Part
}

type Plan struct {
	Changes []Change
	Agents  []AgentPlan
	Kept    []Kept
}

type Planner struct {
	Files  FileStore
	Shared agents.Contribution
	Agents []agents.Agent
	Git    agents.Contribution
	CI     map[string]agents.Contribution
}

type step func(*workspace, context.Context, agents.Contribution) error

func (p Planner) Plan(
	ctx context.Context,
	options agents.Options,
) (Plan, error) {
	return p.run(ctx, options, (*workspace).contribute, false)
}

func (p Planner) Withdraw(
	ctx context.Context,
	options agents.Options,
) (Plan, error) {
	return p.run(ctx, options, (*workspace).withdraw, true)
}

func (p Planner) run(
	ctx context.Context,
	options agents.Options,
	apply step,
	everything bool,
) (Plan, error) {
	space := workspace{files: p.Files, drafts: map[string]*draft{}}
	plan := Plan{}

	if err := apply(&space, ctx, p.Shared); err != nil {
		return Plan{}, err
	}

	for _, agent := range p.Agents {
		contribution, err := agent.Contribute(options)
		if err != nil {
			return Plan{}, err
		}
		if err := apply(&space, ctx, contribution); err != nil {
			return Plan{}, err
		}

		plan.Agents = append(plan.Agents, describe(agent))
	}

	for _, contribution := range p.extras(options, everything) {
		if err := apply(&space, ctx, contribution); err != nil {
			return Plan{}, err
		}
	}

	plan.Changes = space.changes()
	plan.Kept = space.kept

	return plan, nil
}

func (p Planner) extras(
	options agents.Options,
	everything bool,
) []agents.Contribution {
	var found []agents.Contribution

	if options.Precommit {
		found = append(found, p.Git)
	}
	for _, name := range slices.Sorted(maps.Keys(p.CI)) {
		if everything || name == options.CI {
			found = append(found, p.CI[name])
		}
	}

	return found
}

func describe(agent agents.Agent) AgentPlan {
	covers := agent.Moments()
	var missing []agents.Moment

	for _, moment := range agents.Moments() {
		if !slices.Contains(covers, moment) {
			missing = append(missing, moment)
		}
	}

	return AgentPlan{Name: agent.Name(), Covers: covers, Missing: missing}
}

type draft struct {
	before  string
	exists  bool
	after   string
	gone    bool
	touched []agents.Part
}

func (d *draft) set(part agents.Part, after string) {
	if after == d.after {
		return
	}

	d.after = after
	if !slices.Contains(d.touched, part) {
		d.touched = append(d.touched, part)
	}
}

type workspace struct {
	files  FileStore
	drafts map[string]*draft
	order  []string
	kept   []Kept
}

func (w *workspace) load(
	ctx context.Context,
	path string,
) (*draft, error) {
	if current, known := w.drafts[path]; known {
		return current, nil
	}

	text, exists, err := w.files.Read(ctx, path)
	if err != nil {
		return nil, err
	}

	w.drafts[path] = &draft{before: text, exists: exists, after: text}
	w.order = append(w.order, path)

	return w.drafts[path], nil
}

func (w *workspace) contribute(
	ctx context.Context,
	contribution agents.Contribution,
) error {
	for _, file := range contribution.Blocks {
		current, err := w.load(ctx, file.Path)
		if err != nil {
			return err
		}

		if blocks.Covers(current.after, file.Block) {
			continue
		}

		start := current.after
		if start == "" && file.Preamble != "" {
			start = file.Preamble + "\n"
		}

		after, err := blocks.Upsert(start, file.Block)
		if err != nil {
			return err
		}

		current.set(file.Part, after)
	}

	for _, file := range contribution.Entries {
		current, err := w.load(ctx, file.Path)
		if err != nil {
			return err
		}

		after, err := blocks.MergeEntries(current.after, file.Entries)
		if err != nil {
			return err
		}

		current.set(file.Part, after)
	}

	return w.offer(ctx, contribution.Files)
}

func (w *workspace) offer(
	ctx context.Context,
	files []agents.ManagedFile,
) error {
	for _, file := range files {
		current, err := w.load(ctx, file.Path)
		if err != nil {
			return err
		}

		if !ownsFile(current) {
			w.kept = append(w.kept, Kept{file.Path, file.Part})
			continue
		}

		current.set(file.Part, file.Content)
	}

	return nil
}

func ownsFile(current *draft) bool {
	if !current.exists {
		return true
	}

	return blocks.InspectManaged(current.before) == blocks.OwnershipManaged
}

func (w *workspace) withdraw(
	ctx context.Context,
	contribution agents.Contribution,
) error {
	for _, file := range contribution.Blocks {
		current, err := w.load(ctx, file.Path)
		if err != nil {
			return err
		}

		after, status, err := blocks.Remove(current.after, file.Block)
		if err != nil {
			return err
		}
		if status == blocks.RemovalEdited {
			w.kept = append(w.kept, Kept{file.Path, file.Part})
		}

		if file.Preamble != "" && strings.TrimSpace(after) == file.Preamble {
			after = ""
		}

		current.set(file.Part, after)
		current.gone = after == ""
	}

	for _, file := range contribution.Entries {
		current, err := w.load(ctx, file.Path)
		if err != nil {
			return err
		}

		after, left, err := blocks.RemoveEntries(current.after, file.Entries)
		if err != nil {
			return err
		}
		if len(left) > 0 {
			w.kept = append(w.kept, Kept{file.Path, file.Part})
		}

		current.set(file.Part, after)
		current.gone, err = blocks.IsEmptyDocument(after)
		if err != nil {
			return err
		}
	}

	return w.release(ctx, contribution.Files)
}

func (w *workspace) release(
	ctx context.Context,
	files []agents.ManagedFile,
) error {
	for _, file := range files {
		current, err := w.load(ctx, file.Path)
		if err != nil {
			return err
		}

		switch {
		case !current.exists:
			continue
		case !ownsFile(current):
			w.kept = append(w.kept, Kept{file.Path, file.Part})
		default:
			current.set(file.Part, "")
			current.gone = true
		}
	}

	return nil
}

func (w *workspace) changes() []Change {
	var changes []Change

	for _, path := range w.order {
		current := w.drafts[path]
		if current.after == current.before {
			continue
		}

		changes = append(changes, Change{
			Path:    path,
			Created: !current.exists,
			Deleted: current.gone,
			Parts:   current.touched,
			Before:  current.before,
			After:   current.after,
		})
	}

	return changes
}
