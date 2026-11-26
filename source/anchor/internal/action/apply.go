package action

import "context"

type Applier struct {
	Files FileStore
	Stamp string
}

func (a Applier) Apply(ctx context.Context, plan Plan) error {
	for _, change := range plan.Changes {
		if !change.Created {
			err := a.Files.Backup(ctx, a.Stamp, change.Path, change.Before)
			if err != nil {
				return err
			}
		}
	}

	for _, change := range plan.Changes {
		if err := a.put(ctx, change); err != nil {
			return err
		}
	}

	return nil
}

func (a Applier) put(ctx context.Context, change Change) error {
	if change.Deleted {
		return a.Files.Remove(ctx, change.Path)
	}

	return a.Files.Write(ctx, change.Path, change.After)
}
