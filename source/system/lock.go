package system

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"
)

const (
	lockWait    = 5 * time.Second
	lockRetry   = 10 * time.Millisecond
	lockExpired = 30 * time.Second
)

var errLockTimeout = errors.New("timed out waiting for the lock file")

func WithFileLock(path string, work func() error) error {
	lock := path + ".lock"
	if err := acquireLock(lock); err != nil {
		return err
	}

	err := work()
	if released := RemoveFile(lock); err == nil {
		err = released
	}

	return err
}

func acquireLock(lock string) error {
	deadline := time.Now().Add(lockWait)

	for {
		file, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			return file.Close()
		}
		if !isContention(err) {
			return err
		}

		if isExpired(lock) {
			os.Remove(lock)
			continue
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %w", errLockTimeout, err)
		}

		time.Sleep(lockRetry)
	}
}

func isContention(err error) bool {
	return errors.Is(err, fs.ErrExist) || errors.Is(err, fs.ErrPermission)
}

func isExpired(lock string) bool {
	info, err := os.Stat(lock)
	return err == nil && time.Since(info.ModTime()) > lockExpired
}
