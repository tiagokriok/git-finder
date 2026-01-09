package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// GetStatus executes git status and returns formatted output
func GetStatus(repoPath string) (string, error) {
	cmd := exec.Command("git", "status", "--short", "--branch")
	cmd.Dir = repoPath

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git status failed: %s", stderr.String())
	}

	output := out.String()
	if output == "" {
		return "✓ Working tree clean", nil
	}
	return output, nil
}

// GetRemoteURL retrieves the remote origin URL from git config
func GetRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = repoPath

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("no remote configured")
	}

	url := strings.TrimSpace(out.String())
	if url == "" {
		return "", fmt.Errorf("no remote configured")
	}
	return url, nil
}

// ConvertToHTTPS converts SSH git URLs to HTTPS URLs
// Supports formats:
// - git@github.com:user/repo.git
// - ssh://git@github.com/user/repo.git
// - https://github.com/user/repo.git (already HTTPS)
func ConvertToHTTPS(gitURL string) (string, error) {
	gitURL = strings.TrimSpace(gitURL)

	// Already HTTPS
	if strings.HasPrefix(gitURL, "https://") {
		return strings.TrimSuffix(gitURL, ".git"), nil
	}

	// SSH format: git@github.com:user/repo.git
	sshPattern := regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(gitURL); matches != nil {
		host := matches[1]
		path := matches[2]
		return fmt.Sprintf("https://%s/%s", host, path), nil
	}

	// SSH URL format: ssh://git@github.com/user/repo.git
	if strings.HasPrefix(gitURL, "ssh://") {
		urlPattern := regexp.MustCompile(`^ssh://git@([^/]+)/(.+?)(?:\.git)?$`)
		if matches := urlPattern.FindStringSubmatch(gitURL); matches != nil {
			host := matches[1]
			path := matches[2]
			return fmt.Sprintf("https://%s/%s", host, path), nil
		}
	}

	return "", fmt.Errorf("unsupported git URL format: %s", gitURL)
}
