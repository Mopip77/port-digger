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

**logger/** - Logging infrastructure
- Centralized logging system that writes to `~/.config/port-digger/logs/port-digger.log`
- Three log levels: INFO, ERROR, DEBUG
- Special functions for logging lsof queries and LLM requests
- Thread-safe singleton pattern

**scanner/** - Port scanning functionality
- Uses `lsof -iTCP -sTCP:LISTEN -nP` to discover listening TCP ports
- Parses lsof output to extract port number, process name, PID, and protocol
- Returns `PortInfo` structs containing port metadata
- Logs all scan operations for debugging

**actions/** - User actions on ports
- `browser.go`: Opens ports in default browser (http://localhost:PORT)
- `clipboard.go`: Copies port numbers to system clipboard
- `kill.go`: Terminates processes (tries SIGTERM first, falls back to sudo SIGKILL with osascript for password prompt)
- `file.go`: Opens files in default editor

**menu/** - Menu bar UI formatting
- Formats port items for display ("  PORT • ProcessName")
- Supports LLM-rewritten names with ✨ emoji suffix
- Simple utility layer between scanner and systray

**llm/** - LLM integration for process name rewriting
- `config.go`: Configuration management
- `cache.go`: Persistent caching of rewritten names
- `client.go`: API client for LLM requests (logs all requests/responses)
- `rewriter.go`: Orchestration layer with async processing

**main.go** - Application entry point and menu construction
- Initializes systray menu bar app
- Builds dynamic menu from scanned ports
- Wires up action handlers for each port submenu
- Uses goroutines to handle menu click events
- Implements app restart mechanism for refresh functionality

### Key Design Decisions

**On-demand scanning**: The app scans ports when the menu is opened (or refreshed), not continuously in the background. This minimizes resource usage.

**Menu refresh via restart**: Since the systray library doesn't support dynamic menu rebuilding, clicking the Refresh button spawns a new app instance and exits the current one. This ensures the port list is always up-to-date. The same mechanism is used after killing a process.

**Comprehensive logging**: All operations (port scans, LLM requests, user actions) are logged to `~/.config/port-digger/logs/port-digger.log` for debugging and monitoring.

**Privilege escalation**: Process killing uses macOS's native password prompt via `osascript` when sudo is required, providing a secure and familiar UX.

**Clipboard initialization**: Clipboard is initialized once at startup. Errors are non-fatal since the feature can fail gracefully.

**LLM caching**: Process name rewrites are cached persistently to avoid repeated API calls for the same processes.

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
- `llm/*_test.go` - Tests LLM configuration, caching, and client

## Debugging

**View logs in real-time:**
```bash
tail -f ~/.config/port-digger/logs/port-digger.log
```

**Common log entries to look for:**
- `[INFO] Scanning ports...` - Port scan initiated
- `[INFO] lsof query succeeded` - Port scan completed with count
- `[INFO] LLM request succeeded` - LLM rewrite completed
- `[ERROR] ...` - Any errors that occurred
- `[INFO] Refresh button clicked` - User clicked refresh
- `[INFO] Killing process PID X` - User initiated process kill

## Code Patterns

**Error handling**: Non-fatal errors (like clipboard init) print warnings but don't crash. Fatal errors (like lsof failures) return errors to be displayed in the menu.

**Goroutines**: Each menu item spawns a goroutine to handle click events in a select loop, allowing concurrent user interactions.

**Formatting**: Port items are right-aligned with consistent spacing using `fmt.Sprintf("%5d • %s", ...)` for clean visual alignment in the menu.
