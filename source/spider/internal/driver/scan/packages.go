package scan

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

func withWorkspaces(
	root string,
	packages []Package,
	found []finding,
) []Package {
	stack := ""
	for _, item := range found {
		if item.path == "." && isJavascript(item.stack) {
			stack = item.stack
		}
	}

	if stack != "" {
		for _, relative := range workspaceDirectories(root) {
			packages = append(packages, Package{Stack: stack, Path: relative})
		}
	}

	slices.SortFunc(packages, func(a, b Package) int {
		return strings.Compare(a.Path+a.Stack, b.Path+b.Stack)
	})

	return slices.Compact(packages)
}

func isJavascript(stack string) bool {
	return stack == stackBun || stack == stackNode
}

func workspaceDirectories(root string) []string {
	var directories []string

	for _, pattern := range workspacePatterns(root) {
		if strings.HasPrefix(pattern, "!") {
			continue
		}

		pattern = strings.TrimPrefix(path.Clean(pattern), "./")
		matches, _ := filepath.Glob(locate(root, pattern))
		for _, match := range matches {
			relative, err := filepath.Rel(root, match)
			if err == nil && isPackage(match) {
				directories = append(directories, filepath.ToSlash(relative))
			}
		}
	}

	return directories
}

func isPackage(directory string) bool {
	_, err := os.Stat(filepath.Join(directory, "package.json"))

	return err == nil
}

func workspacePatterns(root string) []string {
	var manifest struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}

	patterns := pnpmPatterns(root)
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil || json.Unmarshal(data, &manifest) != nil {
		return patterns
	}

	var list []string
	var object struct {
		Packages []string `json:"packages"`
	}

	if json.Unmarshal(manifest.Workspaces, &list) == nil {
		return append(patterns, list...)
	}
	if json.Unmarshal(manifest.Workspaces, &object) == nil {
		return append(patterns, object.Packages...)
	}

	return patterns
}

func pnpmPatterns(root string) []string {
	data, err := os.ReadFile(filepath.Join(root, "pnpm-workspace.yaml"))
	if err != nil {
		return nil
	}

	var workspace struct {
		Packages []string `yaml:"packages"`
	}
	if yaml.Unmarshal(data, &workspace) != nil {
		return nil
	}

	return workspace.Packages
}
