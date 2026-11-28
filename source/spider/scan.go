package spider

import (
	"github.com/ronz204/beacon/source/spider/internal/driver/scan"
)

type ScanPackage struct {
	Stack string `json:"stack"`
	Path  string `json:"path"`
}

type ScanRunner struct {
	Name  string   `json:"name"`
	Stack string   `json:"stack"`
	Run   []string `json:"run"`
}

type ScanCandidate struct {
	Path  string `json:"path"`
	Files int    `json:"files"`
}

type ScanView struct {
	Stacks     []string        `json:"stacks"`
	Packages   []ScanPackage   `json:"packages"`
	Runners    []ScanRunner    `json:"runners"`
	Candidates []ScanCandidate `json:"candidates"`
}

func Scan(directory string) (ScanView, error) {
	report, err := scan.Scan(directory)
	if err != nil {
		return ScanView{}, err
	}

	view := ScanView{
		Stacks:     report.Stacks,
		Packages:   make([]ScanPackage, 0, len(report.Packages)),
		Runners:    make([]ScanRunner, 0, len(report.Runners)),
		Candidates: make([]ScanCandidate, 0, len(report.Candidates)),
	}
	for _, item := range report.Packages {
		view.Packages = append(view.Packages, ScanPackage(item))
	}
	for _, item := range report.Runners {
		view.Runners = append(view.Runners, ScanRunner(item))
	}
	for _, item := range report.Candidates {
		view.Candidates = append(view.Candidates, ScanCandidate(item))
	}

	return view, nil
}

func (p Project) Scan() (ScanView, error) {
	return Scan(p.root)
}
