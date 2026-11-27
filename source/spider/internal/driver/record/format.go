package record

import (
	"bytes"
	"context"
	"os"

	"github.com/ronz204/beacon/source/kernel"
	"github.com/ronz204/beacon/source/system"
)

type formatTarget struct {
	folder    string
	extension string
	check     func(file string, data []byte) []error
	format    func(file string, data []byte) ([]byte, error)
}

func FormatRecord(file, folder string, data []byte) ([]byte, error) {
	return renderStable(file, normalizeText(data), 0, recordKeyOrders[folder])
}

func FormatConfig(file string, data []byte) ([]byte, error) {
	return renderStable(file, normalizeText(data), 0, configKeys)
}

func FormatGuide(file string, data []byte) ([]byte, error) {
	front, body, found := splitFrontmatter(normalizeText(data))
	if !found {
		return nil, failAt(RecordFrontmatterMissing, file, 1)
	}

	rendered, err := renderStable(file, front, frontmatterOffset, guideKeys)
	if err != nil {
		return nil, err
	}

	text := frontmatterOpen + string(rendered) + frontmatterEnd + body

	return []byte(text), nil
}

func (s Store) Format(ctx context.Context, write bool) ([]string, []error) {
	var changed []string
	var failures []error

	for _, target := range formatTargets() {
		files, listed := listFiles(s.folder(target.folder), target.extension)
		failures = append(failures, listed...)

		for _, file := range files {
			differs, failed := formatFile(ctx, file, target, write)
			failures = append(failures, failed...)

			if differs {
				changed = append(changed, file)
			}
		}
	}

	if _, err := os.Stat(s.ConfigFile()); err == nil {
		file := s.ConfigFile()
		differs, failed := formatFile(ctx, file, configTarget(), write)
		failures = append(failures, failed...)

		if differs {
			changed = append(changed, file)
		}
	}

	return changed, failures
}

func configTarget() formatTarget {
	return formatTarget{
		check: func(file string, data []byte) []error {
			_, failures := decodeConfig(file, data)
			return failures
		},
		format: FormatConfig,
	}
}

func formatTargets() []formatTarget {
	return []formatTarget{
		recordTarget("goals", checkedBy(decodeGoal)),
		recordTarget("capabilities", checkedBy(decodeCapability)),
		recordTarget("decisions", checkedBy(decodeDecision)),
		recordTarget("questions", checkedBy(decodeQuestion)),
		recordTarget("works", checkedBy(decodeWork)),
		recordTarget("tasks", checkedBy(decodeTask)),
		{
			folder:    "guides",
			extension: guideExtension,
			check:     checkedBy(decodeGuide),
			format:    FormatGuide,
		},
	}
}

func recordTarget(
	folder string,
	check func(string, []byte) []error,
) formatTarget {
	return formatTarget{
		folder:    folder,
		extension: recordExtension,
		check:     check,
		format: func(file string, data []byte) ([]byte, error) {
			return FormatRecord(file, folder, data)
		},
	}
}

func checkedBy[T any](decode decoder[T]) func(string, []byte) []error {
	return func(file string, data []byte) []error {
		_, failures := decode(file, data)
		return failures
	}
}

func formatFile(
	ctx context.Context,
	file string,
	target formatTarget,
	write bool,
) (bool, []error) {
	if err := ctx.Err(); err != nil {
		return false, []error{kernel.NewRecordFailure(file, 0, err)}
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return false, []error{kernel.NewRecordFailure(file, 0, err)}
	}

	if failures := target.check(file, data); len(failures) > 0 {
		return false, failures
	}

	output, err := target.format(file, data)
	if err != nil {
		return false, parseFailure(file, err)
	}
	if bytes.Equal(output, data) {
		return false, nil
	}
	if !write {
		return true, nil
	}

	if err := system.WriteFileAtomic(file, output); err != nil {
		return false, []error{kernel.NewRecordFailure(file, 0, err)}
	}

	return true, nil
}
