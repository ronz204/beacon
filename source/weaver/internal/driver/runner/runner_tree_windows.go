//go:build windows

package runner

import (
	"os/exec"
	"strconv"
)

func prepareTree(*exec.Cmd) {}

func killTree(command *exec.Cmd) error {
	pid := strconv.Itoa(command.Process.Pid)

	return exec.Command("taskkill", "/T", "/F", "/PID", pid).Run()
}
