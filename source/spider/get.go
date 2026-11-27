package spider

import (
	"context"
	"os"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/spider/internal/driver/record"
)

func (p Project) Get(ctx context.Context, id string) (StoredRecord, error) {
	parsed, err := kernel.ParseRecordID(id)
	if err != nil {
		return StoredRecord{}, err
	}

	file, err := p.store.Lookup(ctx, parsed)
	if isMissing(err) {
		return p.archived(ctx, id, err)
	}
	if err != nil {
		return StoredRecord{}, err
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return StoredRecord{}, kernel.NewRecordFailure(p.relative(file), 0, err)
	}

	document, err := record.ReadDocument(file, data)
	if err != nil {
		return StoredRecord{}, err
	}

	return StoredRecord{
		Record: Record{Type: document.Kind, Data: document.JSON},
		File:   p.relative(file),
	}, nil
}

func isMissing(err error) bool {
	code := kernel.DescribeFailure(err).Code
	return err != nil && code == string(record.RecordMissing)
}

func (p Project) archived(
	ctx context.Context,
	id string,
	missing error,
) (StoredRecord, error) {
	entry, found, err := p.archive.Find(ctx, id)
	if err != nil {
		return StoredRecord{}, err
	}
	if !found {
		return StoredRecord{}, missing
	}

	return StoredRecord{
		Record:   Record{Type: entry.Type, Data: entry.Record},
		File:     p.relative(entry.File),
		Archived: true,
	}, nil
}
