package weaver

import "context"

func (w Weaver) Author(ctx context.Context) (string, error) {
	person, err := w.records.Identity.Whoami(ctx)
	if err != nil {
		return "", err
	}

	return person.Actor.String(), nil
}
