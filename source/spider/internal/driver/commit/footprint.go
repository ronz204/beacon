package commit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

const deletedMark = "deleted"

func (r Repository) FootprintOf(
	ctx context.Context,
	files []string,
) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	ordered := slices.Clone(files)
	slices.Sort(ordered)
	ordered = slices.Compact(ordered)

	digest := sha256.New()
	for _, file := range ordered {
		if err := ctx.Err(); err != nil {
			return "", kernel.FailHistory(err)
		}

		content, err := r.contentDigest(file)
		if err != nil {
			return "", err
		}

		digest.Write([]byte(file + "\x00" + content + "\n"))
	}

	return hex.EncodeToString(digest.Sum(nil)), nil
}

func (r Repository) contentDigest(file string) (string, error) {
	data, err := os.ReadFile(filepath.Join(r.root, filepath.FromSlash(file)))
	if errors.Is(err, fs.ErrNotExist) {
		return deletedMark, nil
	}
	if err != nil {
		return "", kernel.FailHistory(err)
	}

	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:]), nil
}

func (r Repository) Head(ctx context.Context) (string, error) {
	output, err := r.git(ctx, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}
