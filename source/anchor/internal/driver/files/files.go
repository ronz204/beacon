package files

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/system"
)

const (
	skillsFolder  = "skills"
	stateFolder   = "beacon"
	backupsFolder = "backups"
	directoryMode = 0o755

	hookName        = "hooks/pre-commit"
	hookMode        = 0o755
	StampLayout     = "20060102T150405"
	backupRetention = 30 * 24 * time.Hour
)

type Repo struct {
	root    string
	backups string
	hook    string
}

func Open(ctx context.Context, directory string) (Repo, error) {
	top, err := system.Git(ctx, directory, "rev-parse", "--show-toplevel")
	if err != nil {
		return Repo{}, kernel.FailHistory(err)
	}

	root := filepath.Clean(strings.TrimSpace(top))
	gitDir, err := system.GitCommonDir(ctx, root)
	if err != nil {
		return Repo{}, kernel.FailHistory(err)
	}

	backups := filepath.Join(gitDir, stateFolder, backupsFolder)

	hook, err := system.Git(ctx, root, "rev-parse", "--git-path", hookName)
	if err != nil {
		return Repo{}, kernel.FailHistory(err)
	}

	hook = filepath.FromSlash(strings.TrimSpace(hook))
	if !filepath.IsAbs(hook) {
		hook = filepath.Join(root, hook)
	}

	return Repo{root: root, backups: backups, hook: hook}, nil
}

func (r Repo) HookPath() string {
	relative, err := filepath.Rel(r.root, r.hook)
	if err == nil && filepath.IsLocal(relative) {
		return filepath.ToSlash(relative)
	}

	return filepath.ToSlash(r.hook)
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
	if err == nil && target == r.hook {
		err = os.Chmod(target, hookMode)
	}
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

	dropEmptySkillFolder(target)

	return nil
}

func dropEmptySkillFolder(file string) {
	folder := filepath.Dir(file)
	if filepath.Base(filepath.Dir(folder)) == skillsFolder {
		_ = os.Remove(folder)
	}
}

func (r Repo) Backup(
	_ context.Context,
	stamp, path, content string,
) error {
	folder := filepath.Join(r.backups, stamp)
	if filepath.IsAbs(filepath.FromSlash(path)) {
		path = hookName
	}

	target, err := r.resolve(folder, path)
	if err != nil {
		return err
	}

	err = writeFile(target, content)
	if err != nil {
		return action.FailInstall(action.BackupFailed, path, err)
	}

	r.pruneBackups(time.Now())

	return nil
}

func (r Repo) pruneBackups(now time.Time) {
	entries, err := os.ReadDir(r.backups)
	if err != nil {
		return
	}

	for _, entry := range entries {
		taken, err := time.ParseInLocation(
			StampLayout,
			entry.Name(),
			time.Local,
		)
		if err == nil && now.Sub(taken) > backupRetention {
			_ = os.RemoveAll(filepath.Join(r.backups, entry.Name()))
		}
	}
}

func (r Repo) resolve(base, path string) (string, error) {
	local := filepath.FromSlash(path)

	if base == r.root && local == r.hook {
		return local, nil
	}
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
