package main

import (
	"os/exec"
	"strings"
)

// getGitInfo returns the current short commit SHA and branch name.
func getGitInfo() (commitSHA, branch string) {
	commitSHA = "unknown"

	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	branchOut, err := exec.Command("git", "branch", "--show-current").Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	return commitSHA, branch
}
