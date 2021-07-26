package mages

import (
	"fmt"

	"github.com/magefile/mage/sh"
	"github.com/pubgo/xerror"
)

// GitHash returns the git hash for the current repo or "" if none.
func GitHash(n int) string {
	hash, err := sh.Output("git", "rev-parse", fmt.Sprintf("--short=%d", n), "HEAD")
	xerror.Exit(err)
	return hash
}

// GitTag returns the git tag for the current branch or "" if none.
func GitTag() string {
	//git tag --sort=committerdate | tail -n 1
	s, err := sh.Output("git", "tag", "--sort=committerdate", "|", "tail -n 1")
	//s, err := sh.Output("git", "describe", "--abbrev=0", "--tags")
	xerror.Exit(err)
	return s
}
