package files

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/system"
)

const (
	stateFolder   = "beacon"
	backupsFolder = "backups"
	directoryMode = 0o755
)

type Repo struct {
	root    string
	backups string
}

func Open(ctx context.Context, directory string) (Repo, error) {
	top, err := system.Git(ctx, directory, "rev-parse", "--show-toplevel")
	if err != nil {
		return Repo{}, kernel.FailHistory(err)
	}

	root := filepath.Clean(strings.TrimSpace(top))
	common, err := system.Git(ctx, root, "rev-parse", "--git-common-dir")
	if err != nil {
		return Repo{}, kernel.FailHistory(err)
	}

	gitDir := strings.TrimSpace(common)
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(root, gitDir)
	}

	backups := filepath.Join(gitDir, stateFolder, backupsFolder)

	return Repo{root: root, backups: backups}, nil
}

func (r Repo) Read(
	_ context.Context,
	path string,
) (string, bool, error) {
	target, err := r.resolve(r.root, path)
	if err != nil {
		return "", false, err
	}

	data, err := os.ReadFile(target)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, action.FailInstall(action.FileReadFailed, path, err)
	}

	return string(data), true, nil
}

func (r Repo) Write(_ context.Context, path, content string) error {
	target, err := r.resolve(r.root, path)
	if err != nil {
		return err
	}

	err = writeFile(target, content)
	if err != nil {
		return action.FailInstall(action.FileWriteFailed, path, err)
	}

	return nil
}

func (r Repo) Remove(_ context.Context, path string) error {
	target, err := r.resolve(r.root, path)
	if err != nil {
		return err
	}

	err = system.RemoveFile(target)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return action.FailInstall(action.FileRemoveFailed, path, err)
	}

	return nil
}

func (r Repo) Backup(
	_ context.Context,
	stamp, path, content string,
) error {
	folder := filepath.Join(r.backups, stamp)

	target, err := r.resolve(folder, path)
	if err != nil {
		return err
	}

	err = writeFile(target, content)
	if err != nil {
		return action.FailInstall(action.BackupFailed, path, err)
	}

	return nil
}

func (r Repo) resolve(base, path string) (string, error) {
	local := filepath.FromSlash(path)

	if !filepath.IsLocal(local) {
		return "", action.FailInstall(action.FilePathInvalid, path, nil)
	}

	return filepath.Join(base, local), nil
}

func writeFile(target, content string) error {
	err := os.MkdirAll(filepath.Dir(target), directoryMode)
	if err != nil {
		return err
	}

	return system.WriteFileAtomic(target, []byte(content))
}
