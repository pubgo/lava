package buildtasks

import (
	"os/exec"

	"github.com/goyek/goyek/v2"
	"github.com/goyek/x/cmd"
)

// check if docker is installed and running
func checkDocker(a *goyek.A) bool {
	if !hasBinary("docker") {
		return true
	}

	return cmd.Exec(a, "docker ps")
}

// check if a binary exists
func hasBinary(binaryName string) bool {
	_, err := exec.LookPath(binaryName)
	return err == nil
}
