package scan

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const maxDepth = 3

var skipped = []string{
	"node_modules", "vendor", "dist", "build", "target", "__pycache__",
}

type Package struct {
	Stack string
	Path  string
}

type Runner struct {
	Name  string
	Stack string
	Run   []string
}

type Candidate struct {
	Path  string
	Files int
}

type Report struct {
	Stacks     []string
	Packages   []Package
	Runners    []Runner
	Candidates []Candidate
}

type finding struct {
	stack   string
	manager string
	path    string
}

func Scan(root string) (Report, error) {
	var found []finding
	if err := walk(root, "", 0, &found); err != nil {
		return Report{}, err
	}

	report := Report{
		Stacks:   []string{},
		Packages: []Package{},
		Runners:  []Runner{},
	}
	managers := map[string]string{}

	for _, item := range found {
		if !slices.Contains(report.Stacks, item.stack) {
			report.Stacks = append(report.Stacks, item.stack)
		}
		if _, known := managers[item.stack]; !known {
			managers[item.stack] = item.manager
		}

		report.Packages = append(report.Packages, Package{
			Stack: item.stack,
			Path:  item.path,
		})
	}

	slices.Sort(report.Stacks)
	report.Runners = runnersFor(report.Stacks, managers)
	report.Packages = withWorkspaces(root, report.Packages, found)

	candidates, err := candidatesOf(root, report.Packages)
	report.Candidates = candidates

	return report, err
}

func walk(root, relative string, depth int, found *[]finding) error {
	entries, err := os.ReadDir(locate(root, relative))
	if err != nil {
		return err
	}

	names := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			names[entry.Name()] = true
		}
	}

	for _, stack := range detect(names, depth == 0) {
		location := relative
		if location == "" {
			location = "."
		}

		*found = append(*found, finding{
			stack:   stack.name,
			manager: stack.manager,
			path:    location,
		})
	}

	if depth == maxDepth {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() && !ignored(entry.Name()) {
			child := path.Join(relative, entry.Name())
			if err := walk(root, child, depth+1, found); err != nil {
				return err
			}
		}
	}

	return nil
}

func locate(root, relative string) string {
	return filepath.Join(root, filepath.FromSlash(relative))
}

func ignored(name string) bool {
	return strings.HasPrefix(name, ".") || slices.Contains(skipped, name)
}
