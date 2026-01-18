# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Port Digger is a lightweight macOS menu bar application written in Go that monitors TCP listening ports in real-time. It provides a simple interface to view, interact with, and manage processes listening on ports.

## Build & Development Commands

```bash
# Build the application
make build

# Build optimized release version
make build-release

# Run tests
make test

# Run the application (builds first)
make run

# Install to ~/Applications
make install

# Clean build artifacts
make clean
```

## Running Individual Tests

```bash
# Run tests for a specific package
go test ./scanner -v
go test ./actions -v
go test ./menu -v

# Run a specific test function
go test ./scanner -run TestScanPorts -v
```

## Architecture

Port Digger follows a modular architecture with clear separation of concerns:

### Core Components

**scanner/** - Port scanning functionality
- Uses `lsof -iTCP -sTCP:LISTEN -nP` to discover listening TCP ports
- Parses lsof output to extract port number, process name, PID, and protocol
- Returns `PortInfo` structs containing port metadata

**actions/** - User actions on ports
- `browser.go`: Opens ports in default browser (http://localhost:PORT)
- `clipboard.go`: Copies port numbers to system clipboard
- `kill.go`: Terminates processes (tries SIGTERM first, falls back to sudo SIGKILL with osascript for password prompt)

**menu/** - Menu bar UI formatting
- Formats port items for display ("  PORT • ProcessName")
- Simple utility layer between scanner and systray

**main.go** - Application entry point and menu construction
- Initializes systray menu bar app
- Builds dynamic menu from scanned ports
- Wires up action handlers for each port submenu
- Uses goroutines to handle menu click events

### Key Design Decisions

**On-demand scanning**: The app scans ports when the menu is opened (or refreshed), not continuously in the background. This minimizes resource usage.

**Menu limitations**: The systray library doesn't support dynamic menu rebuilding easily. The Refresh button is present but has limitations in updating the menu structure.

**Privilege escalation**: Process killing uses macOS's native password prompt via `osascript` when sudo is required, providing a secure and familiar UX.

**Clipboard initialization**: Clipboard is initialized once at startup. Errors are non-fatal since the feature can fail gracefully.

## Dependencies

- `github.com/getlantern/systray` - Menu bar integration
- `github.com/skratchdot/open-golang` - Cross-platform browser opening
- `golang.design/x/clipboard` - System clipboard access
- Go 1.24.5+

## Testing Strategy

Tests use standard Go testing patterns. Each package has its own `_test.go` files:
- `scanner/scanner_test.go` - Tests lsof parsing and port scanning
- `actions/*_test.go` - Tests browser, clipboard, and kill actions
- `menu/builder_test.go` - Tests menu formatting

## Code Patterns

**Error handling**: Non-fatal errors (like clipboard init) print warnings but don't crash. Fatal errors (like lsof failures) return errors to be displayed in the menu.

**Goroutines**: Each menu item spawns a goroutine to handle click events in a select loop, allowing concurrent user interactions.

**Formatting**: Port items are right-aligned with consistent spacing using `fmt.Sprintf("%5d • %s", ...)` for clean visual alignment in the menu.
