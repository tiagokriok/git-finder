package platform

import (
	"fmt"
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
		// Try common Linux terminal emulators in order
		for _, term := range []string{
			"gnome-terminal",
			"konsole",
			"xfce4-terminal",
			"alacritty",
			"kitty",
			"wezterm",
			"xterm",
		} {
			if _, err := exec.LookPath(term); err == nil {
				return term
			}
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
