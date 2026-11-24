package system_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ronz204/beacon/internal/system"
)

func TestWithFileLockSerializesWriters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "counter.txt")
	if err := os.WriteFile(path, []byte("0"), 0o644); err != nil {
		t.Fatal(err)
	}

	const writers = 20
	var group sync.WaitGroup
	for range writers {
		group.Go(func() {
			err := system.WithFileLock(path, func() error {
				return increment(path)
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
	group.Wait()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != strconv.Itoa(writers) {
		t.Fatalf("counter = %s, want %d", data, writers)
	}
	if _, err := os.Stat(path + ".lock"); err == nil {
		t.Fatal("lock file left behind")
	}
}

func increment(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}

	next := []byte(strconv.Itoa(value + 1))

	return system.WriteFileAtomic(path, next)
}

func TestWithFileLockRecoversAnExpiredLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	lock := path + ".lock"
	if err := os.WriteFile(lock, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lock, past, past); err != nil {
		t.Fatal(err)
	}

	ran := false
	err := system.WithFileLock(path, func() error {
		ran = true
		return nil
	})

	if err != nil || !ran {
		t.Fatalf("err = %v, ran = %v", err, ran)
	}
}
