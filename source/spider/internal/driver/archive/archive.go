package archive

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/domain/thread"
	"github.com/ronz204/beacon/source/system"
)

var errUnknownType = errors.New("unknown archive record type")

const (
	segmentExtension = ".jsonl"
	promiseKind      = "promise"
	maximumLine      = 16 << 20
)

var nodeKinds = map[string]thread.NodeKind{
	"goal":       thread.NodeGoal,
	"capability": thread.NodeCapability,
	promiseKind:  thread.NodePromise,
	"decision":   thread.NodeDecision,
	"question":   thread.NodeQuestion,
	"work":       thread.NodeWork,
	"task":       thread.NodeTask,
}

type Entry struct {
	ID     string
	Type   string
	Record json.RawMessage
	File   string
	Line   int
}

type Store struct {
	folder string
}

type lineDocument struct {
	Type   string          `json:"type"`
	Record json.RawMessage `json:"record"`
}

type recordIDs struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Promises   []struct {
		ID string `json:"id"`
	} `json:"promises"`
}

func NewStore(folder string) Store {
	return Store{folder: folder}
}

func SegmentName(closed kernel.CalendarDate) string {
	return closed.String()[:7] + segmentExtension
}

func (s Store) Entries(ctx context.Context) ([]Entry, error) {
	files, err := s.segments()
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		read, err := readSegment(file)
		if err != nil {
			return nil, err
		}

		entries = append(entries, read...)
	}

	return entries, nil
}

func (s Store) segments() ([]string, error) {
	items, err := os.ReadDir(s.folder)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, kernel.NewRecordFailure(s.folder, 0, err)
	}

	var files []string
	for _, item := range items {
		if !item.IsDir() && filepath.Ext(item.Name()) == segmentExtension {
			files = append(files, filepath.Join(s.folder, item.Name()))
		}
	}
	slices.Sort(files)

	return files, nil
}

func readSegment(file string) ([]Entry, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, kernel.NewRecordFailure(file, 0, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(nil, maximumLine)

	var entries []Entry
	for number := 1; scanner.Scan(); number++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}

		var line lineDocument
		if err := json.Unmarshal([]byte(text), &line); err != nil {
			return nil, kernel.NewRecordFailure(file, number, err)
		}

		var ids recordIDs
		_ = json.Unmarshal(line.Record, &ids)

		entries = append(entries, Entry{
			ID:     ids.ID,
			Type:   line.Type,
			Record: line.Record,
			File:   file,
			Line:   number,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, kernel.NewRecordFailure(file, 0, err)
	}

	return entries, nil
}

func (s Store) ArchivedNodes(ctx context.Context) ([]thread.GraphNode, error) {
	entries, err := s.Entries(ctx)
	if err != nil {
		return nil, err
	}

	var nodes []thread.GraphNode
	for _, entry := range entries {
		read, err := entryNodes(entry)
		if err != nil {
			return nil, kernel.NewRecordFailure(entry.File, entry.Line, err)
		}

		nodes = append(nodes, read...)
	}

	return nodes, nil
}

func entryNodes(entry Entry) ([]thread.GraphNode, error) {
	kind, known := nodeKinds[entry.Type]
	if !known {
		return nil, errUnknownType
	}

	var ids recordIDs
	if err := json.Unmarshal(entry.Record, &ids); err != nil {
		return nil, err
	}

	id, err := kernel.ParseRecordID(ids.ID)
	if err != nil {
		return nil, err
	}

	node := thread.GraphNode{ID: id, Kind: kind}
	if kind == thread.NodePromise {
		if node.Owner, err = kernel.ParseRecordID(ids.Capability); err != nil {
			return nil, err
		}
	}

	nodes := []thread.GraphNode{node}
	for _, promise := range ids.Promises {
		promiseID, err := kernel.ParseRecordID(promise.ID)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, thread.GraphNode{
			ID:    promiseID,
			Kind:  thread.NodePromise,
			Owner: id,
		})
	}

	return nodes, nil
}

func (s Store) Find(ctx context.Context, id string) (Entry, bool, error) {
	entries, err := s.Entries(ctx)
	if err != nil {
		return Entry{}, false, err
	}

	for _, entry := range entries {
		if entry.ID == id {
			return entry, true, nil
		}
	}

	return Entry{}, false, nil
}

func (s Store) Append(name string, kind string, record []byte) error {
	file := filepath.Join(s.folder, name)

	current, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return kernel.NewRecordFailure(file, 0, err)
	}

	var line bytes.Buffer
	line.Write(current)
	if len(current) > 0 && !bytes.HasSuffix(current, []byte{'\n'}) {
		line.WriteByte('\n')
	}
	kindJSON, _ := json.Marshal(kind)
	line.WriteString(`{"type":`)
	line.Write(kindJSON)
	line.WriteString(`,"record":`)
	line.Write(record)
	line.WriteString("}\n")

	if err := os.MkdirAll(s.folder, 0o755); err != nil {
		return kernel.NewRecordFailure(file, 0, err)
	}
	if err := system.WriteFileAtomic(file, line.Bytes()); err != nil {
		return kernel.NewRecordFailure(file, 0, err)
	}

	return nil
}
