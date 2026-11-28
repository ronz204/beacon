package weaver

import (
	"context"
	"strings"
)

const humanPrefix = "human:"

func (w Weaver) PersonPresent(ctx context.Context) bool {
	person, err := w.records.Identity.Whoami(ctx)

	return err == nil && strings.HasPrefix(person.Actor.String(), humanPrefix)
}
