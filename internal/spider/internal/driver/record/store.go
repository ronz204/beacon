package record

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ronz204/beacon/internal/kernel"
	"github.com/ronz204/beacon/internal/spider/internal/domain/agenda"
	"github.com/ronz204/beacon/internal/spider/internal/domain/ledger"
	"github.com/ronz204/beacon/internal/spider/internal/domain/memory"
)

const (
	recordExtension = ".yaml"
	guideExtension  = ".md"
)

type Store struct {
	root string
}

type decoder[T any] func(file string, data []byte) (T, []error)

type outcome[T any] struct {
	value    T
	failures []error
}

func NewStore(root string) Store {
	return Store{root: root}
}

func (s Store) ListGoals(ctx context.Context) ([]ledger.Goal, []error) {
	return loadFolder(ctx, s.folder("goals"), recordExtension, decodeGoal)
}

func (s Store) ListCapabilities(
	ctx context.Context,
) ([]ledger.Capability, []error) {
	folder := s.folder("capabilities")
	return loadFolder(ctx, folder, recordExtension, decodeCapability)
}

func (s Store) ListDecisions(
	ctx context.Context,
) ([]memory.Decision, []error) {
	folder := s.folder("decisions")
	return loadFolder(ctx, folder, recordExtension, decodeDecision)
}

func (s Store) ListQuestions(
	ctx context.Context,
) ([]memory.Question, []error) {
	folder := s.folder("questions")
	return loadFolder(ctx, folder, recordExtension, decodeQuestion)
}

func (s Store) ListGuides(ctx context.Context) ([]memory.Guide, []error) {
	return loadFolder(ctx, s.folder("guides"), guideExtension, decodeGuide)
}

func (s Store) ListWorks(ctx context.Context) ([]agenda.Work, []error) {
	return loadFolder(ctx, s.folder("works"), recordExtension, decodeWork)
}

func (s Store) ListTasks(ctx context.Context) ([]agenda.Task, []error) {
	return loadFolder(ctx, s.folder("tasks"), recordExtension, decodeTask)
}

func (s Store) folder(name string) string {
	return filepath.Join(s.root, "records", name)
}

func loadFolder[T any](
	ctx context.Context,
	folder string,
	extension string,
	decode decoder[T],
) ([]T, []error) {
	files, failures := listFiles(folder, extension)
	outcomes := make([]outcome[T], len(files))
	jobs := make(chan int)

	var workers sync.WaitGroup
	for range min(runtime.GOMAXPROCS(0), len(files)) {
		workers.Go(func() {
			for index := range jobs {
				outcomes[index] = loadFile(ctx, files[index], decode)
			}
		})
	}

	for index := range files {
		jobs <- index
	}
	close(jobs)
	workers.Wait()

	values := make([]T, 0, len(files))
	for _, result := range outcomes {
		failures = append(failures, result.failures...)

		if len(result.failures) == 0 {
			values = append(values, result.value)
		}
	}

	return values, failures
}

func listFiles(folder, extension string) ([]string, []error) {
	entries, err := os.ReadDir(folder)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, []error{kernel.NewRecordFailure(folder, 0, err)}
	}

	var files []string
	var failures []error

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(folder, name)

		switch {
		case entry.IsDir(), strings.HasPrefix(name, "."):
			continue
		case filepath.Ext(name) != extension:
			cause := failAt(RecordFileExtension, path, 0)
			failures = append(failures, kernel.NewRecordFailure(path, 0, cause))
		default:
			files = append(files, path)
		}
	}

	return files, failures
}

func loadFile[T any](
	ctx context.Context,
	file string,
	decode decoder[T],
) outcome[T] {
	if err := ctx.Err(); err != nil {
		failure := kernel.NewRecordFailure(file, 0, err)
		return outcome[T]{failures: []error{failure}}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		failure := kernel.NewRecordFailure(file, 0, err)
		return outcome[T]{failures: []error{failure}}
	}

	value, failures := decode(file, data)

	return outcome[T]{value: value, failures: failures}
}
