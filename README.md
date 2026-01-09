# GF - Git Fuzzy

A high-performance CLI tool for discovering and opening Git repositories with interactive fuzzy search. Built with Go, leveraging modern terminal UI patterns and efficient filesystem traversal.

## Key Features

- **Real-time Fuzzy Search**: Instant repository filtering with substring matching
- **Interactive TUI**: Responsive terminal UI using The Elm Architecture (Bubbletea)
- **Intelligent Scanning**: Recursive filesystem traversal with automatic exclusion of non-essential directories
- **Editor Integration**: Seamless handoff to configured editor (Neovim, VS Code, Vim, etc.)
- **Zero-Configuration Setup**: Interactive wizard creates sensible defaults on first run
- **Performance Optimized**: Efficient directory traversal with early termination and deduplication

## Quick Start

### Installation

**Requirements**: Go 1.25.5 or higher

#### Option 1: Using `go install` (Fastest) ⭐

```bash
go install github.com/tiagokriok/Git-Fuzzy/cmd/gitf@latest
gitf
```

This installs directly to `$GOPATH/bin` (ensure `$GOPATH/bin` is in your `$PATH`).

#### Option 2: Using Makefile (Recommended for Development)

```bash
git clone https://github.com/tiagokriok/Git-Fuzzy.git
cd Git-Fuzzy
make build          # Build the binary
make install        # Install to $GOPATH/bin
```

#### Option 3: Manual Build from Source

##### Linux
```bash
git clone https://github.com/tiagokriok/Git-Fuzzy.git
cd Git-Fuzzy
go build -o gitf ./cmd/gitf
sudo mv gitf /usr/local/bin/  # Install system-wide
# OR
mv gitf ~/.local/bin/         # Install user-local
```

##### macOS
```bash
git clone https://github.com/tiagokriok/Git-Fuzzy.git
cd Git-Fuzzy
go build -o gitf ./cmd/gitf
sudo mv gitf /usr/local/bin/  # Install system-wide
# OR using Homebrew path
mv gitf /usr/local/opt/gitf/bin/
```

##### Windows (PowerShell)
```powershell
git clone https://github.com/tiagokriok/Git-Fuzzy.git
cd Git-Fuzzy
go build -o gitf.exe ./cmd/gitf
Move-Item gitf.exe $env:GOPATH\bin\  # Install to GOPATH/bin
```

### Initial Setup & Configuration Wizard

On first run, GF automatically launches an **interactive setup wizard** that:

1. **Creates config directory**: `~/.config/gitf/` (Linux/macOS) or `%APPDATA%\gitf\` (Windows)
2. **Prompts for editor selection**: Enter your preferred editor command
3. **Prompts for repository paths**: Enter directories to scan for Git repositories

#### Running the Setup Wizard

Simply run `gitf` for the first time:

```bash
gitf
```

The wizard will interactively guide you through configuration. It's implemented in `internal/ui/setup.go`.

#### Example Setup Wizard Flow

```
Welcome to Git Fuzzy Setup!

Enter your preferred editor (nvim, vim, code, etc.): nvim

Enter repository search paths (one per line, empty line to finish):
Path 1: ~/dev
Path 2: ~/projects
Path 3: ~/work
Path 4: [empty - wizard finishes]

Config saved to ~/.config/gitf/config.json
```

### How to Configure Repository Paths

Repository paths tell GF where to scan for Git repositories. You can configure them in two ways:

#### Method 1: Interactive Setup Wizard

Run the wizard again at any time:
```bash
gitf --setup
```

Or simply delete the config and run `gitf`:
```bash
rm ~/.config/gitf/config.json
gitf
```

#### Method 2: Direct Config File Editing

Edit `~/.config/gitf/config.json` directly:

**Linux/macOS:**
```bash
nano ~/.config/gitf/config.json
```

**Windows (PowerShell):**
```powershell
notepad $env:APPDATA\gitf\config.json
```

### Configuration File Examples

#### Linux/macOS Example
```json
{
  "editor": "nvim",
  "search_paths": [
    "/home/user/dev",
    "/home/user/projects",
    "/home/user/work",
    "/opt/repositories"
  ]
}
```

**Common macOS paths:**
```json
{
  "editor": "code",
  "search_paths": [
    "~/Developer",
    "~/Projects",
    "~/workspace",
    "/Volumes/external-drive/repos"
  ]
}
```

#### Windows Example (PowerShell format)
```json
{
  "editor": "code",
  "search_paths": [
    "C:\\Users\\YourUsername\\dev",
    "C:\\Users\\YourUsername\\projects",
    "C:\\workspace",
    "D:\\repositories"
  ]
}
```

Or using backslash escaping in JSON:
```json
{
  "editor": "code.exe",
  "search_paths": [
    "C:\\Users\\YourUsername\\source\\repos",
    "C:\\work",
    "E:\\projects"
  ]
}
```

### Basic Usage

After configuration, simply run:

```bash
gitf
```

GF will:
1. Scan all configured `search_paths` for Git repositories
2. Display them in an interactive terminal UI
3. Let you filter by typing (fuzzy search)
4. Open your selection in the configured editor

**Note**: GF intelligently skips common directories like `node_modules`, `vendor`, `.git`, and virtual environments to ensure fast scanning even in large codebases.

### Makefile Commands

```bash
make help              # Show all available commands
make build             # Build the gitf binary
make build-optimized   # Build optimized binary (29% smaller)
make run               # Run gitf directly
make test              # Run all tests
make test-verbose      # Run tests with verbose output
make test-coverage     # Generate coverage report (opens HTML)
make fmt               # Format code
make lint              # Run go vet
make clean             # Remove build artifacts
make install           # Install to $GOPATH/bin
make deps              # Download and manage dependencies
make reset-local       # Reset config and binary for fresh start
make check             # Run fmt, lint, and test
make dev               # Full development workflow (clean, fmt, lint, test, build)
```

## Keyboard Shortcuts

### Main View

- `↑` / `↓` or `Tab` / `Shift+Tab`: Navigate repositories
- `Type`: Filter by repository name (fuzzy search)
- `Backspace`: Delete character from search
- `Enter`: Open selected repository in editor
- `Ctrl+O`: Open file manager at repository location
- `Ctrl+T`: Open terminal in repository directory
- `Ctrl+B`: Open remote repository in browser (GitHub/GitLab)
- `Ctrl+G`: Show git status in modal overlay
- `Esc` / `Ctrl+C`: Exit application

### Git Status Modal

- `↑` / `↓` or `j` / `k`: Scroll through git status
- `Esc` or `q`: Close modal and return to main view

## Technical Architecture

### Project Structure

```
gf/
├── cmd/gitf/
│   └── main.go                      # Application entry point & editor integration
├── internal/
│   ├── config/
│   │   ├── config.go                # Configuration management (load/save/defaults)
│   │   └── config_test.go           # Unit tests (62.9% coverage)
│   ├── history/
│   │   └── recent.go                # Recent repositories tracking & persistence
│   ├── scanner/
│   │   ├── scanner.go               # Repository discovery with optimized traversal
│   │   │                             # Includes ReorderByRecent() for intelligent sorting
│   │   └── scanner_test.go          # Unit tests (60.0% coverage)
│   └── ui/
│       ├── ui.go                    # Interactive TUI & result rendering
│       └── setup.go                 # Setup wizard for first-run configuration
├── Makefile                         # Development workflow automation
├── .gitignore                       # Git ignore patterns
├── go.mod
└── go.sum
```

### Design Patterns

- **Separation of Concerns**: Modular architecture with config, scanner, and UI packages
- **The Elm Architecture**: TUI implementation using Bubbletea for predictable state management
- **Efficient Traversal**: `filepath.Walk()` with early `SkipDir` for large directory trees
- **Deduplication**: Map-based tracking prevents duplicate repository entries
- **Error Handling**: Comprehensive error wrapping with context using `fmt.Errorf("%w")`

## Development

### Build & Test

```bash
# Build
go build -o gitf ./cmd/gitf

# Run tests with verbose output
go test -v ./...

# Check test coverage
go test -cover ./...

# Run specific test
go test -v -run TestNameHere ./internal/config
```

### Code Quality

```bash
# Format code (Go standard)
gofmt -l -w .
go fmt ./...

# Static analysis
go vet ./...

# Update dependencies
go mod tidy
go mod download
```

### Test Coverage Summary

- **Config Package**: 62.9% coverage (6 tests)
- **Scanner Package**: 60.0% coverage (5 tests)
- **History Package**: Implemented (pending unit tests)
- **UI Package**: Manual testing (TUI interaction testing)
- **Total**: 11 passing tests ✅

## Technology Stack

### Core Dependencies

| Package | Purpose | Version |
|---------|---------|---------|
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework implementing Elm Architecture | Latest |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling and layout | Latest |
| [sahilm/fuzzy](https://github.com/sahilm/fuzzy) | Efficient fuzzy string matching | Latest |

### Go Version
- **Go 1.25.5** or higher required

## Configuration

Configuration is stored in `~/.config/gitf/config.json` (Linux/macOS) or `%APPDATA%\gitf\config.json` (Windows) and created automatically on first run via the setup wizard.

### Configuration Options

| Option | Type | Description | Example |
|--------|------|-------------|---------|
| `editor` | string | Command to launch when opening repository | `"nvim"`, `"code"`, `"vim"`, `"code.exe"` (Windows) |
| `search_paths` | array | Directories to recursively scan for Git repos | `["/home/user/dev", "/work"]` or `["C:\\\\Users\\\\user\\\\dev"]` |

### Configuration File Locations

| OS | Default Location |
|----|------------------|
| **Linux** | `~/.config/gitf/config.json` |
| **macOS** | `~/.config/gitf/config.json` |
| **Windows** | `%APPDATA%\gitf\config.json` |

### How to Edit Configuration

**Option 1: Use the setup wizard**
```bash
gitf --setup  # Re-run interactive wizard
```

**Option 2: Manual edit (Linux/macOS)**
```bash
nano ~/.config/gitf/config.json
vim ~/.config/gitf/config.json
```

**Option 3: Manual edit (Windows PowerShell)**
```powershell
notepad $env:APPDATA\gitf\config.json
```

### Performance Optimization

The scanner automatically skips these directories to improve traversal speed:

```
node_modules, vendor, target        # Dependency/build directories
.git, .config, .vscode, .idea       # Hidden configuration directories
.cache, venv, .venv, venv3, .venv3  # Cache and virtual environments
```

This intelligent filtering enables sub-second repository discovery even in large directory trees.

## Recent Updates

### Version 0.1.0

#### ✨ New Features
- **Makefile**: Complete development workflow automation with 15+ targets
- **Binary Optimization**: Build-time flag optimization reduces binary size by 29% (5.1M → 3.6M)
- **Recent Repositories**: Track and prioritize last 10 opened repositories
- **Intelligent Sorting**: Automatically reorder search results by most recent usage

#### 🔧 Improvements
- Test Coverage: Now at 11 passing tests across config and scanner packages
- Code Quality: Full `go fmt`, `go vet` compliance with error wrapping

#### 📊 Project Stats
- **Lines of Code**: ~990 (production: 700, tests: 290)
- **Binary Size**: 5.1M (standard), 3.6M (optimized)
- **Test Pass Rate**: 100% (11/11 tests passing)
- **Supported Go Version**: 1.25.5+

## License

MIT
