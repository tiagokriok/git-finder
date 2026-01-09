package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tiagokriok/Git-Fuzzy/internal/config"
	"github.com/tiagokriok/Git-Fuzzy/internal/history"
	"github.com/tiagokriok/Git-Fuzzy/internal/scanner"
	"github.com/tiagokriok/Git-Fuzzy/internal/ui"
)

// Version information (injected at build time via ldflags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "source"
)

func getVersionInfo() string {
	return fmt.Sprintf("gitf version %s\ncommit: %s\nbuilt at: %s\nbuilt by: %s",
		version, commit, date, builtBy)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "gitf",
		Short: "Git Fuzzy - Interactive repository selector",
		Long: `Git Fuzzy (gitf) is a fast, interactive git repository finder.

It scans configured directories for git repositories and lets you
quickly select one using fuzzy search. Selected repositories open
automatically in your preferred editor.`,
		Version:      getVersionInfo(),
		RunE:         runTUI,
		SilenceUsage: true,
	}

	// Add --setup flag
	var setupFlag bool
	rootCmd.Flags().BoolVarP(&setupFlag, "setup", "s", false,
		"Force configuration wizard to run")

	// Handle --setup before main RunE
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if setupFlag {
			if err := handleSetup(cmd, args); err != nil {
				return err
			}
			os.Exit(0)
		}
		return nil
	}

	// Custom version template for cleaner output
	rootCmd.SetVersionTemplate(`{{.Version}}` + "\n")

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg, err = ui.RunSetup()
			if err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	repos, err := scanner.Scan(cfg.SearchPaths)
	if err != nil {
		return fmt.Errorf("failed to scan repositories: %w", err)
	}

	recents, err := history.LoadRecent()
	if err == nil {
		repos = scanner.ReorderByRecent(repos, recents)
	}

	if len(repos) == 0 {
		return fmt.Errorf("no repositories found in search paths")
	}

	selected, err := ui.Run(repos)
	if err != nil {
		return fmt.Errorf("failed to run UI: %w", err)
	}

	if selected == nil {
		return nil // User cancelled
	}

	if err := openRepositoryInEditor(cfg.Editor, selected.Path); err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	recent, err := history.LoadRecent()
	if err == nil {
		recent.Add(selected.Path)
		recent.Save()
	}

	return nil
}

func handleSetup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg != nil {
		fmt.Println("\nCurrent configuration:")
		fmt.Printf("  Editor: %s\n", cfg.Editor)
		fmt.Printf("  Search Paths: %s\n\n", strings.Join(cfg.SearchPaths, ", "))
		fmt.Println("Press Enter to setup or Ctrl+C to cancel...")
		fmt.Scanln()
	}

	newCfg, err := ui.RunSetup()
	if err != nil {
		return fmt.Errorf("setup cancelled: %w", err)
	}

	if err := newCfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("\n✓ Configuration saved successfully!")
	return nil
}

func openRepositoryInEditor(editor string, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
