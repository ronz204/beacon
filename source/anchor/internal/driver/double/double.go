package double

import (
	"context"
	"path"
)

type Files struct {
	Texts   map[string]string
	Backups map[string]string
	Writes  []string
	Err     error
}

func NewFiles(texts map[string]string) *Files {
	return &Files{Texts: texts, Backups: map[string]string{}}
}

func (f *Files) Read(
	_ context.Context,
	name string,
) (string, bool, error) {
	if f.Err != nil {
		return "", false, f.Err
	}

	text, exists := f.Texts[name]

	return text, exists, nil
}

func (f *Files) Write(_ context.Context, name, content string) error {
	if f.Err != nil {
		return f.Err
	}

	f.Texts[name] = content
	f.Writes = append(f.Writes, name)

	return nil
}

func (f *Files) Backup(
	_ context.Context,
	stamp, name, content string,
) error {
	if f.Err != nil {
		return f.Err
	}

	f.Backups[path.Join(stamp, name)] = content

	return nil
}
