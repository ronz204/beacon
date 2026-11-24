package system

import (
	"errors"
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
	defer os.Remove(lock)

	return work()
}

func acquireLock(lock string) error {
	deadline := time.Now().Add(lockWait)

	for {
		file, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			return file.Close()
		}
		if !errors.Is(err, fs.ErrExist) {
			return err
		}

		if isExpired(lock) {
			os.Remove(lock)
			continue
		}
		if time.Now().After(deadline) {
			return errLockTimeout
		}

		time.Sleep(lockRetry)
	}
}

func isExpired(lock string) bool {
	info, err := os.Stat(lock)
	return err == nil && time.Since(info.ModTime()) > lockExpired
}
