package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// StatusData contains detailed git status information
type StatusData struct {
	CurrentBranch    string
	TrackingBranch   string
	AheadCount       int
	BehindCount      int
	StashCount       int
	ModifiedCount    int
	AddedCount       int
	DeletedCount     int
	RenamedCount     int
	CopiedCount      int
	UntrackedCount   int
	Files            []FileStatus
}

// FileStatus represents a single file's git status
type FileStatus struct {
	Status   string // M, A, D, ??, R, C, etc.
	Filename string
}

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

// GetDetailedStatus returns comprehensive git status information
func GetDetailedStatus(repoPath string) (*StatusData, error) {
	data := &StatusData{
		Files: make([]FileStatus, 0),
	}

	// Get current branch and tracking info
	if err := getCurrentBranchInfo(repoPath, data); err != nil {
		return nil, err
	}

	// Get ahead/behind counts
	if data.TrackingBranch != "" {
		if err := getAheadBehindCounts(repoPath, data); err != nil {
			// Non-fatal error, continue
		}
	}

	// Get stash count
	if err := getStashCount(repoPath, data); err != nil {
		// Non-fatal error, continue
	}

	// Parse file status
	if err := parseFileStatus(repoPath, data); err != nil {
		return nil, err
	}

	return data, nil
}

// getCurrentBranchInfo gets the current branch and tracking branch
func getCurrentBranchInfo(repoPath string, data *StatusData) error {
	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	data.CurrentBranch = strings.TrimSpace(out.String())

	// Get tracking branch
	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	cmd.Dir = repoPath
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		tracking := strings.TrimSpace(out.String())
		if tracking != "@{u}" && tracking != "" {
			data.TrackingBranch = tracking
		}
	}

	return nil
}

// getAheadBehindCounts gets how many commits ahead/behind the tracking branch
func getAheadBehindCounts(repoPath string, data *StatusData) error {
	// Use rev-list to count commits
	cmd := exec.Command("git", "rev-list", "--count", "--left-right",
		fmt.Sprintf("%s...HEAD", data.TrackingBranch))
	cmd.Dir = repoPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}

	parts := strings.Fields(strings.TrimSpace(out.String()))
	if len(parts) == 2 {
		if behind, err := strconv.Atoi(parts[0]); err == nil {
			data.BehindCount = behind
		}
		if ahead, err := strconv.Atoi(parts[1]); err == nil {
			data.AheadCount = ahead
		}
	}

	return nil
}

// getStashCount gets the number of stashed changes
func getStashCount(repoPath string, data *StatusData) error {
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = repoPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) > 0 && lines[0] != "" {
		data.StashCount = len(lines)
	}

	return nil
}

// parseFileStatus parses git status output and categorizes files
func parseFileStatus(repoPath string, data *StatusData) error {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = repoPath
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to get file status: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse status line: "XY filename"
		// First two characters are status codes
		if len(line) < 3 {
			continue
		}

		status := line[:2]
		filename := strings.TrimSpace(line[3:])

		// Count by type
		countFileStatus(status, data)

		// Add to file list
		data.Files = append(data.Files, FileStatus{
			Status:   strings.TrimSpace(status),
			Filename: filename,
		})
	}

	return nil
}

// countFileStatus counts a file by its status type
func countFileStatus(status string, data *StatusData) {
	status = strings.TrimSpace(status)

	switch {
	case strings.Contains(status, "M"):
		data.ModifiedCount++
	case strings.Contains(status, "A"):
		data.AddedCount++
	case strings.Contains(status, "D"):
		data.DeletedCount++
	case strings.Contains(status, "R"):
		data.RenamedCount++
	case strings.Contains(status, "C"):
		data.CopiedCount++
	case status == "??":
		data.UntrackedCount++
	}
}
