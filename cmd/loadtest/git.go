package main

import (
	"context"
	"os/exec"
	"strings"
)

// getGitInfo returns the current short commit SHA and branch name.
func getGitInfo() (string, string) {
	ctx := context.Background()
	commitSHA := "unknown"

	out, err := exec.CommandContext(ctx, "git", "rev-parse", "--short", "HEAD").Output()
	if err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	branch := ""

	branchOut, err := exec.CommandContext(ctx, "git", "branch", "--show-current").Output()
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	return commitSHA, branch
}
