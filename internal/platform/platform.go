package platform

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// DetectFileManager returns the default file manager command for the current platform
func DetectFileManager() string {
	switch runtime.GOOS {
	case "linux":
		// Try common Linux file managers in order
		for _, fm := range []string{"nautilus", "dolphin", "thunar", "nemo", "caja"} {
			if _, err := exec.LookPath(fm); err == nil {
				return fm
			}
		}
		// Fallback to xdg-open
		if _, err := exec.LookPath("xdg-open"); err == nil {
			return "xdg-open"
		}
		return ""
	case "darwin":
		// macOS uses 'open' command
		return "open"
	case "windows":
		// Windows uses explorer
		return "explorer"
	default:
		return ""
	}
}

// DetectTerminal returns the default terminal command for the current platform
func DetectTerminal() string {
	switch runtime.GOOS {
	case "linux":
		// Use xdg-terminal-exec for default terminal (modern standard)
		if _, err := exec.LookPath("xdg-terminal-exec"); err == nil {
			return "xdg-terminal-exec"
		}
		// Fallback to x-terminal-emulator (Debian/Ubuntu)
		if _, err := exec.LookPath("x-terminal-emulator"); err == nil {
			return "x-terminal-emulator"
		}
		return ""
	case "darwin":
		// macOS: use Terminal.app via open command
		return "open -a Terminal"
	case "windows":
		// Windows: Check for PowerShell, WSL, then cmd
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell"
		}
		if _, err := exec.LookPath("wsl"); err == nil {
			return "wsl"
		}
		if _, err := exec.LookPath("cmd"); err == nil {
			return "cmd"
		}
		return ""
	default:
		return ""
	}
}

// DetectCurrentTerminal detects the terminal emulator the app is currently running in
// by checking terminal-specific environment variables
func DetectCurrentTerminal() string {
	// Check terminal-specific environment variables
	terminalChecks := []struct {
		envVar   string
		terminal string
	}{
		// Ghostty
		{"GHOSTTY_RESOURCES_DIR", "ghostty"},
		// Kitty
		{"KITTY_WINDOW_ID", "kitty"},
		// Alacritty
		{"ALACRITTY_SOCKET", "alacritty"},
		{"ALACRITTY_LOG", "alacritty"},
		// WezTerm
		{"WEZTERM_PANE", "wezterm"},
		{"WEZTERM_EXECUTABLE", "wezterm"},
		// iTerm2 (macOS)
		{"ITERM_SESSION_ID", "iterm2"},
		// Hyper
		{"HYPER_TERM", "hyper"},
		// Terminology
		{"TERMINOLOGY", "terminology"},
		// Tilix
		{"TILIX_ID", "tilix"},
		// Konsole
		{"KONSOLE_VERSION", "konsole"},
		// GNOME Terminal
		{"GNOME_TERMINAL_SCREEN", "gnome-terminal"},
		// xterm / generic
		{"XTERM_VERSION", "xterm"},
	}

	for _, check := range terminalChecks {
		if os.Getenv(check.envVar) != "" {
			// Verify the terminal is actually installed
			if _, err := exec.LookPath(check.terminal); err == nil {
				return check.terminal
			}
		}
	}

	// Check TERM_PROGRAM (used by many terminals)
	if termProgram := os.Getenv("TERM_PROGRAM"); termProgram != "" {
		// Normalize common values
		switch termProgram {
		case "Apple_Terminal":
			return "open -a Terminal"
		case "iTerm.app":
			return "open -a iTerm"
		case "vscode":
			// VS Code integrated terminal - fall through to system detection
		default:
			// Try using it directly if it's in PATH
			if _, err := exec.LookPath(termProgram); err == nil {
				return termProgram
			}
		}
	}

	// Check user-set TERMINAL env var
	if terminal := os.Getenv("TERMINAL"); terminal != "" {
		if _, err := exec.LookPath(terminal); err == nil {
			return terminal
		}
	}

	// Fallback to system detection
	return ""
}

// OpenInBrowser opens a URL in the default browser
func OpenInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		// Try xdg-open on Linux
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		// Use 'open' on macOS
		cmd = exec.Command("open", url)
	case "windows":
		// Use 'start' command on Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Fire and forget (non-blocking)
	return cmd.Start()
}
