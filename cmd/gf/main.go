package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/tiagokriok/gf/internal/config"
	"github.com/tiagokriok/gf/internal/scanner"
	"github.com/tiagokriok/gf/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	repos, err := scanner.Scan(cfg.SearchPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning repositories: %v\n", err)
		os.Exit(1)
	}

	if len(repos) == 0 {
		fmt.Fprintf(os.Stderr, "No repositories found\n")
		os.Exit(1)
	}

	selected, err := ui.Run(repos)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
		os.Exit(1)
	}

	if selected == nil {
		os.Exit(0)
	}

	err = openRepositoryInEditor(cfg.Editor, selected.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository in editor: %v\n", err)
		os.Exit(1)
	}

}

func openRepositoryInEditor(editor string, path string) error {
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
