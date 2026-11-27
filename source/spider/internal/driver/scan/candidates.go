package scan

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

var containers = []string{
	"src", "source", "internal", "pkg", "packages", "apps", "lib",
	"services", "modules", "cmd",
}

var notCode = []string{
	"docs", "doc", "test", "tests", "testing", "scripts", "assets",
	"public", "examples", "fixtures",
}

var codeExtensions = []string{
	".go", ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".py", ".rs",
	".java", ".kt", ".rb", ".php", ".cs", ".c", ".cpp", ".h", ".swift",
	".vue", ".svelte",
}

func candidatesOf(root string, packages []Package) ([]Candidate, error) {
	folders := workspaceFolders(packages)
	if len(folders) == 0 {
		var err error
		folders, err = topFolders(root, "")
		if err != nil {
			return nil, err
		}
	}

	candidates := []Candidate{}
	for _, folder := range folders {
		files, err := countFiles(locate(root, folder))
		if err != nil {
			return nil, err
		}
		if files > 0 {
			candidates = append(candidates, Candidate{
				Path:  folder,
				Files: files,
			})
		}
	}

	slices.SortFunc(candidates, func(a, b Candidate) int {
		if a.Files != b.Files {
			return b.Files - a.Files
		}

		return strings.Compare(a.Path, b.Path)
	})

	return candidates, nil
}

func workspaceFolders(packages []Package) []string {
	var folders []string

	for _, item := range packages {
		if item.Path != "." && !slices.Contains(folders, item.Path) {
			folders = append(folders, item.Path)
		}
	}

	if len(folders) == 0 {
		return nil
	}

	return folders
}

func topFolders(root, relative string) ([]string, error) {
	entries, err := os.ReadDir(locate(root, relative))
	if err != nil {
		return nil, err
	}

	var folders []string
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || ignored(name) || skippedFolder(name) {
			continue
		}

		child := path.Join(relative, name)
		if !slices.Contains(containers, name) {
			folders = append(folders, child)
			continue
		}

		inner, err := topFolders(root, child)
		if err != nil {
			return nil, err
		}

		folders = append(folders, inner...)
	}

	return folders, nil
}

func skippedFolder(name string) bool {
	return slices.Contains(notCode, name)
}

func isCode(file string) bool {
	return slices.Contains(codeExtensions, filepath.Ext(file))
}

func countFiles(directory string) (int, error) {
	count := 0

	err := filepath.WalkDir(directory, func(
		current string,
		entry fs.DirEntry,
		err error,
	) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && current != directory && ignored(entry.Name()) {
			return filepath.SkipDir
		}
		if !entry.IsDir() && isCode(current) {
			count++
		}

		return nil
	})

	return count, err
}
