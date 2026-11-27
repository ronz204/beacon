package catalog

import (
	"bytes"
	"errors"
	"io/fs"
	"path"
	"slices"
	"strings"
	"text/template"

	"github.com/ronz204/beacon/source/anchor/internal/action"
	"github.com/ronz204/beacon/source/anchor/internal/agents"
)

const (
	rulesFolder    = "rules"
	skillsFolder   = "skills"
	skillFile      = "SKILL.md"
	templateSuffix = ".tmpl"
	ruleSuffix     = ".md"
)

type Language struct {
	Code string
	UI   string
}

type Commit struct {
	Min int
	Max int
}

type Binding struct {
	Language Language
	Commit   Commit
}

func DefaultBinding() Binding {
	return Binding{
		Language: Language{Code: "en"},
		Commit:   Commit{Min: 40, Max: 50},
	}
}

type Catalog struct {
	source fs.FS
}

func New(source fs.FS) Catalog {
	return Catalog{source: source}
}

func (c Catalog) Everything(binding Binding) ([]agents.Piece, error) {
	packs, err := c.packNames()
	if err != nil {
		return nil, err
	}

	return c.piecesOf(packs, binding)
}

func (c Catalog) Select(
	packs, skip []string,
	binding Binding,
) ([]agents.Piece, error) {
	known, err := c.packNames()
	if err != nil {
		return nil, err
	}

	for _, pack := range packs {
		if !slices.Contains(known, pack) {
			return nil, action.FailPackName(action.PackUnknown, pack)
		}
	}

	every, err := c.piecesOf(known, binding)
	if err != nil {
		return nil, err
	}

	for _, name := range skip {
		if !hasPiece(every, name) {
			return nil, action.FailPackName(action.PieceUnknown, name)
		}
	}

	chosen, err := c.piecesOf(packs, binding)
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(chosen, func(piece agents.Piece) bool {
		return slices.Contains(skip, piece.Name)
	}), nil
}

func hasPiece(pieces []agents.Piece, name string) bool {
	return slices.ContainsFunc(pieces, func(piece agents.Piece) bool {
		return piece.Name == name
	})
}

func (c Catalog) packNames() ([]string, error) {
	entries, err := fs.ReadDir(c.source, ".")
	if err != nil {
		return nil, assetFailure(".", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	return names, nil
}

func (c Catalog) piecesOf(
	packs []string,
	binding Binding,
) ([]agents.Piece, error) {
	var pieces []agents.Piece

	for _, pack := range packs {
		rules, err := c.rules(pack, binding)
		if err != nil {
			return nil, err
		}

		skills, err := c.skills(pack, binding)
		if err != nil {
			return nil, err
		}

		pieces = append(pieces, rules...)
		pieces = append(pieces, skills...)
	}

	return pieces, nil
}

func (c Catalog) rules(
	pack string,
	binding Binding,
) ([]agents.Piece, error) {
	folder := path.Join(pack, rulesFolder)

	entries, err := readFolder(c.source, folder)
	if err != nil {
		return nil, err
	}

	var pieces []agents.Piece
	for _, entry := range entries {
		file := strings.TrimSuffix(entry.Name(), templateSuffix)
		name, isRule := strings.CutSuffix(file, ruleSuffix)
		if entry.IsDir() || !isRule {
			continue
		}

		content, err := c.render(path.Join(folder, entry.Name()), binding)
		if err != nil {
			return nil, err
		}

		pieces = append(pieces, agents.Piece{
			Pack:    pack,
			Kind:    agents.PieceRule,
			Name:    name,
			Content: content,
		})
	}

	return pieces, nil
}

func (c Catalog) skills(
	pack string,
	binding Binding,
) ([]agents.Piece, error) {
	folder := path.Join(pack, skillsFolder)

	entries, err := readFolder(c.source, folder)
	if err != nil {
		return nil, err
	}

	var pieces []agents.Piece
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		content, err := c.renderSkill(
			path.Join(folder, entry.Name()),
			binding,
		)
		if err != nil {
			return nil, err
		}

		pieces = append(pieces, agents.Piece{
			Pack:    pack,
			Kind:    agents.PieceSkill,
			Name:    entry.Name(),
			Content: content,
		})
	}

	return pieces, nil
}

func (c Catalog) renderSkill(
	folder string,
	binding Binding,
) (string, error) {
	target := path.Join(folder, skillFile)

	if _, err := fs.Stat(c.source, target+templateSuffix); err == nil {
		return c.render(target+templateSuffix, binding)
	}

	return c.render(target, binding)
}

func (c Catalog) render(target string, binding Binding) (string, error) {
	data, err := fs.ReadFile(c.source, target)
	if err != nil {
		return "", assetFailure(target, err)
	}

	text := string(data)
	if !strings.HasSuffix(target, templateSuffix) {
		return text, nil
	}

	parsed, err := template.New(target).
		Option("missingkey=error").
		Parse(text)
	if err != nil {
		return "", assetFailure(target, err)
	}

	var rendered bytes.Buffer
	if err := parsed.Execute(&rendered, binding); err != nil {
		return "", assetFailure(target, err)
	}

	return rendered.String(), nil
}

func readFolder(source fs.FS, folder string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(source, folder)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, assetFailure(folder, err)
	}

	return entries, nil
}

func assetFailure(target string, cause error) error {
	return action.FailInstall(action.PackAssetInvalid, target, cause)
}
