package benchmark

import (
	"encoding/json"
	"os/exec"
)

func Inspect(directory string, task Task, beacon Beacon) (Outcome, error) {
	if err := copyTree(task.HiddenPath(), directory); err != nil {
		return Outcome{}, err
	}

	outcome := Outcome{Resolved: hiddenTestsPass(directory, task)}

	text, err := beacon(directory, "diff", "--json")
	if err != nil {
		return outcome, err
	}

	var reply struct {
		Data struct {
			Pending []json.RawMessage `json:"pending"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(text), &reply); err != nil {
		return outcome, err
	}

	outcome.Pending = len(reply.Data.Pending)

	return outcome, nil
}

func hiddenTestsPass(directory string, task Task) bool {
	if len(task.Hidden.Run) == 0 {
		return false
	}

	command := exec.Command(task.Hidden.Run[0], task.Hidden.Run[1:]...)
	command.Dir = directory

	return command.Run() == nil
}
