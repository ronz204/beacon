package benchmark

import (
	"context"
	"io"
	"os"
)

type Agent interface {
	Run(ctx context.Context, directory string, task Task) (io.ReadCloser, error)
}

type Replay struct {
	Transcript string
}

func (r Replay) Run(
	_ context.Context,
	_ string,
	_ Task,
) (io.ReadCloser, error) {
	return os.Open(r.Transcript)
}
