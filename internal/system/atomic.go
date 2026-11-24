package system

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultFileMode fs.FileMode = 0o644
	renameAttempts              = 5
	renameBackoff               = 10 * time.Millisecond
)

func WriteFileAtomic(path string, data []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".beacon-*")
	if err != nil {
		return err
	}

	name := temporary.Name()
	defer os.Remove(name)

	if err := writeAndClose(temporary, data); err != nil {
		return err
	}
	if err := os.Chmod(name, targetMode(path)); err != nil {
		return err
	}

	return renameWithRetry(name, path)
}

func writeAndClose(file *os.File, data []byte) error {
	_, err := file.Write(data)
	if err == nil {
		err = file.Sync()
	}

	if closeErr := file.Close(); err == nil {
		err = closeErr
	}

	return err
}

func targetMode(path string) fs.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return defaultFileMode
	}

	return info.Mode().Perm()
}

func renameWithRetry(source, target string) error {
	var err error

	for attempt := range renameAttempts {
		err = os.Rename(source, target)
		if err == nil {
			return nil
		}

		time.Sleep(time.Duration(attempt+1) * renameBackoff)
	}

	return err
}
