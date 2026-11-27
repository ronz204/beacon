package benchmark

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

type Condition string

const (
	ConditionWithout    Condition = "A"
	ConditionGuide      Condition = "B"
	ConditionFull       Condition = "C"
	ConditionNoReinject Condition = "D"
)

type Beacon func(directory string, args ...string) (string, error)

func Conditions() []Condition {
	return []Condition{
		ConditionWithout,
		ConditionGuide,
		ConditionFull,
		ConditionNoReinject,
	}
}

func (c Condition) InstallArgs() ([]string, error) {
	switch c {
	case ConditionWithout:
		return nil, nil
	case ConditionGuide:
		return []string{"install", "--skip", "hooks"}, nil
	case ConditionFull:
		return []string{"install"}, nil
	case ConditionNoReinject:
		return []string{"install", "--skip", "reinject"}, nil
	default:
		return nil, fmt.Errorf("unknown condition %q", c)
	}
}

func Prepare(
	directory string,
	task Task,
	condition Condition,
	beacon Beacon,
) error {
	args, err := condition.InstallArgs()
	if err != nil {
		return err
	}

	if err := copyTree(task.TemplatePath(), directory); err != nil {
		return err
	}

	steps := [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.name", "jane"},
		{"config", "user.email", "jane@example.com"},
		{"add", "."},
		{"commit", "-q", "-m", "template"},
	}
	for _, step := range steps {
		if err := git(directory, step...); err != nil {
			return err
		}
	}

	if args == nil {
		return nil
	}

	if _, err := beacon(directory, args...); err != nil {
		return err
	}

	return git(directory, "add", ".")
}

func git(directory string, args ...string) error {
	command := exec.Command("git", args...)
	command.Dir = directory

	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v: %w: %s", args, err, output)
	}

	return nil
}

func copyTree(source, target string) error {
	return filepath.WalkDir(source, func(
		path string,
		entry fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destination, data, 0o644)
	})
}
