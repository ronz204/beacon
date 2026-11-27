package bench

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Kind      string `json:"kind"`
	Request   string `json:"request"`
	Template  string `json:"template"`
	Tempts    string `json:"tempts,omitempty"`
	Hidden    Hidden `json:"hidden"`
	Directory string `json:"-"`
}

type Hidden struct {
	Dir string   `json:"dir"`
	Run []string `json:"run"`
}

func LoadTask(file string) (Task, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Task{}, err
	}

	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return Task{}, err
	}

	task.Directory = filepath.Dir(file)

	return task, nil
}

func (t Task) TemplatePath() string {
	return filepath.Join(t.Directory, filepath.FromSlash(t.Template))
}

func (t Task) HiddenPath() string {
	return filepath.Join(t.Directory, filepath.FromSlash(t.Hidden.Dir))
}
