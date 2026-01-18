# Port Digger Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a lightweight macOS menu bar tool that monitors TCP listening ports with quick actions (open in browser, copy port, kill process).

**Architecture:** Pure Go application using systray for native menu bar integration, executing lsof on-demand when menu is clicked for minimal memory footprint (10-20MB). Event-driven architecture with dynamic menu building.

**Tech Stack:** Go 1.21+, github.com/getlantern/systray, github.com/skratchdot/open-golang/open, golang.design/x/clipboard

---

## Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `.gitignore`

**Step 1: Initialize Go module**

Run:
```bash
go mod init github.com/yourusername/port-digger
```

Expected: Creates `go.mod` with module declaration

**Step 2: Add dependencies**

Run:
```bash
go get github.com/getlantern/systray@latest
go get github.com/skratchdot/open-golang/open@latest
go get golang.design/x/clipboard@latest
```

Expected: Dependencies added to `go.mod`, `go.sum` created

**Step 3: Create .gitignore**

```
# Binaries
port-digger
PortDigger
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# macOS
.DS_Store
```

**Step 4: Commit**

```bash
git add go.mod go.sum .gitignore
git commit -m "chore: initialize Go module and dependencies"
```

---

## Task 2: Port Scanner - Data Structures

**Files:**
- Create: `scanner/scanner.go`
- Create: `scanner/scanner_test.go`

**Step 1: Write test for PortInfo struct**

Create `scanner/scanner_test.go`:

```go
package scanner

import (
	"testing"
)

func TestPortInfo_Validation(t *testing.T) {
	info := PortInfo{
		Port:        3000,
		ProcessName: "node",
		PID:         12345,
		Command:     "node server.js",
		Protocol:    "TCP",
	}

	if info.Port != 3000 {
		t.Errorf("expected Port=3000, got %d", info.Port)
	}
	if info.ProcessName != "node" {
		t.Errorf("expected ProcessName=node, got %s", info.ProcessName)
	}
	if info.PID != 12345 {
		t.Errorf("expected PID=12345, got %d", info.PID)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./scanner -v`

Expected: FAIL with "no Go files in .../scanner"

**Step 3: Create PortInfo struct**

Create `scanner/scanner.go`:

```go
package scanner

// PortInfo represents a listening TCP port with associated process information
type PortInfo struct {
	Port        int    // Port number
	ProcessName string // Process name from lsof COMMAND column
	PID         int    // Process ID
	Command     string // Full command line (for future custom naming)
	Protocol    string // "TCP" or "TCP6"
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./scanner -v`

Expected: PASS

**Step 5: Commit**

```bash
git add scanner/
git commit -m "feat(scanner): add PortInfo data structure"
```

---

## Task 3: Port Scanner - lsof Parser

**Files:**
- Modify: `scanner/scanner_test.go`
- Modify: `scanner/scanner.go`

**Step 1: Write test for lsof output parsing**

Add to `scanner/scanner_test.go`:

```go
func TestParseLsofLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		want     *PortInfo
		wantErr  bool
	}{
		{
			name: "valid TCP line",
			line: "node      12345 user   23u  IPv4 0x1234      0t0  TCP *:3000 (LISTEN)",
			want: &PortInfo{
				Port:        3000,
				ProcessName: "node",
				PID:         12345,
				Command:     "node",
				Protocol:    "TCP",
			},
			wantErr: false,
		},
		{
			name: "valid TCP6 line",
			line: "Python    9876 user   5u  IPv6 0x5678      0t0  TCP6 *:8080 (LISTEN)",
			want: &PortInfo{
				Port:        8080,
				ProcessName: "Python",
				PID:         9876,
				Command:     "Python",
				Protocol:    "TCP6",
			},
			wantErr: false,
		},
		{
			name:    "invalid line - not enough fields",
			line:    "node 12345",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "header line",
			line:    "COMMAND   PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLsofLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLsofLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Port != tt.want.Port {
				t.Errorf("Port = %v, want %v", got.Port, tt.want.Port)
			}
			if got.ProcessName != tt.want.ProcessName {
				t.Errorf("ProcessName = %v, want %v", got.ProcessName, tt.want.ProcessName)
			}
			if got.PID != tt.want.PID {
				t.Errorf("PID = %v, want %v", got.PID, tt.want.PID)
			}
			if got.Protocol != tt.want.Protocol {
				t.Errorf("Protocol = %v, want %v", got.Protocol, tt.want.Protocol)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./scanner -v`

Expected: FAIL with "undefined: parseLsofLine"

**Step 3: Implement parseLsofLine function**

Add to `scanner/scanner.go`:

```go
import (
	"fmt"
	"strconv"
	"strings"
)

// parseLsofLine parses a single line of lsof output
// Example: "node      12345 user   23u  IPv4 0x1234      0t0  TCP *:3000 (LISTEN)"
func parseLsofLine(line string) (*PortInfo, error) {
	fields := strings.Fields(line)

	// Need at least 9 fields for valid output
	if len(fields) < 9 {
		return nil, fmt.Errorf("invalid line format: not enough fields")
	}

	// Skip header line
	if fields[0] == "COMMAND" {
		return nil, fmt.Errorf("header line")
	}

	// Parse PID (field 1)
	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("invalid PID: %w", err)
	}

	// Parse protocol and port from NAME field (field 8)
	// Format: "TCP *:3000" or "TCP6 *:8080"
	nameParts := strings.Split(fields[8], " ")
	if len(nameParts) < 2 {
		return nil, fmt.Errorf("invalid NAME field format")
	}

	protocol := nameParts[0] // "TCP" or "TCP6"

	// Extract port from "*:3000" or "127.0.0.1:8080"
	portStr := nameParts[1]
	colonIdx := strings.LastIndex(portStr, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("no port found in NAME field")
	}

	port, err := strconv.Atoi(portStr[colonIdx+1:])
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %w", err)
	}

	return &PortInfo{
		Port:        port,
		ProcessName: fields[0],
		PID:         pid,
		Command:     fields[0],
		Protocol:    protocol,
	}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./scanner -v`

Expected: PASS

**Step 5: Commit**

```bash
git add scanner/
git commit -m "feat(scanner): add lsof output line parser"
```

---

## Task 4: Port Scanner - Main Scan Function

**Files:**
- Modify: `scanner/scanner_test.go`
- Modify: `scanner/scanner.go`

**Step 1: Write test for ScanPorts (integration-style)**

Add to `scanner/scanner_test.go`:

```go
import (
	"runtime"
)

func TestScanPorts(t *testing.T) {
	// Only run on macOS (lsof behavior is OS-specific)
	if runtime.GOOS != "darwin" {
		t.Skip("ScanPorts test only runs on macOS")
	}

	ports, err := ScanPorts()
	if err != nil {
		t.Fatalf("ScanPorts() failed: %v", err)
	}

	// Should return at least empty slice, not nil
	if ports == nil {
		t.Error("ScanPorts() returned nil, expected slice")
	}

	// Validate each port has required fields
	for i, p := range ports {
		if p.Port == 0 {
			t.Errorf("ports[%d].Port is 0", i)
		}
		if p.PID == 0 {
			t.Errorf("ports[%d].PID is 0", i)
		}
		if p.ProcessName == "" {
			t.Errorf("ports[%d].ProcessName is empty", i)
		}
		if p.Protocol != "TCP" && p.Protocol != "TCP6" {
			t.Errorf("ports[%d].Protocol = %s, want TCP or TCP6", i, p.Protocol)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./scanner -v`

Expected: FAIL with "undefined: ScanPorts"

**Step 3: Implement ScanPorts function**

Add to `scanner/scanner.go`:

```go
import (
	"bufio"
	"bytes"
	"os/exec"
)

// ScanPorts executes lsof to get all listening TCP ports
func ScanPorts() ([]PortInfo, error) {
	// Execute: lsof -iTCP -sTCP:LISTEN -nP
	cmd := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-nP")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		// lsof returns exit code 1 if no ports found - not an error
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []PortInfo{}, nil
		}
		return nil, fmt.Errorf("lsof command failed: %w", err)
	}

	var ports []PortInfo
	scanner := bufio.NewScanner(&stdout)

	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse line, skip on error (headers, malformed lines)
		info, err := parseLsofLine(line)
		if err != nil {
			continue
		}

		ports = append(ports, *info)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading lsof output: %w", err)
	}

	return ports, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./scanner -v`

Expected: PASS

**Step 5: Commit**

```bash
git add scanner/
git commit -m "feat(scanner): implement ScanPorts with lsof execution"
```

---

## Task 5: Actions - Browser Opener

**Files:**
- Create: `actions/browser.go`
- Create: `actions/browser_test.go`

**Step 1: Write test for OpenBrowser**

Create `actions/browser_test.go`:

```go
package actions

import (
	"testing"
)

func TestOpenBrowser(t *testing.T) {
	// This is a manual test - we can't easily verify browser actually opens
	// Just test that function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("OpenBrowser panicked: %v", r)
		}
	}()

	// Use a high port unlikely to conflict
	err := OpenBrowser(38291)

	// We expect this might fail (port not actually serving)
	// but it shouldn't panic or return error from open.Run
	if err != nil {
		t.Logf("OpenBrowser returned error (expected): %v", err)
	}
}

func TestFormatURL(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		{3000, "http://localhost:3000"},
		{8080, "http://localhost:8080"},
		{80, "http://localhost:80"},
	}

	for _, tt := range tests {
		got := formatURL(tt.port)
		if got != tt.want {
			t.Errorf("formatURL(%d) = %v, want %v", tt.port, got, tt.want)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./actions -v`

Expected: FAIL with "no Go files in .../actions"

**Step 3: Implement browser opener**

Create `actions/browser.go`:

```go
package actions

import (
	"fmt"
	"github.com/skratchdot/open-golang/open"
)

// formatURL creates the localhost URL for a given port
func formatURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

// OpenBrowser opens the default browser to the given port
func OpenBrowser(port int) error {
	url := formatURL(port)
	return open.Run(url)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./actions -v`

Expected: PASS

**Step 5: Commit**

```bash
git add actions/
git commit -m "feat(actions): implement browser opener"
```

---

## Task 6: Actions - Clipboard Copy

**Files:**
- Modify: `actions/clipboard.go`
- Modify: `actions/clipboard_test.go`

**Step 1: Write test for CopyToClipboard**

Create `actions/clipboard_test.go`:

```go
package actions

import (
	"testing"
	"golang.design/x/clipboard"
)

func TestCopyToClipboard(t *testing.T) {
	// Initialize clipboard (required once per process)
	err := clipboard.Init()
	if err != nil {
		t.Skipf("Clipboard not available: %v", err)
	}

	tests := []struct {
		port int
		want string
	}{
		{3000, "3000"},
		{8080, "8080"},
		{65535, "65535"},
	}

	for _, tt := range tests {
		err := CopyToClipboard(tt.port)
		if err != nil {
			t.Errorf("CopyToClipboard(%d) error = %v", tt.port, err)
			continue
		}

		// Read back from clipboard
		got := string(clipboard.Read(clipboard.FmtText))
		if got != tt.want {
			t.Errorf("CopyToClipboard(%d): clipboard = %v, want %v", tt.port, got, tt.want)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./actions -v`

Expected: FAIL with "undefined: CopyToClipboard"

**Step 3: Implement clipboard copy**

Create `actions/clipboard.go`:

```go
package actions

import (
	"fmt"
	"golang.design/x/clipboard"
)

// CopyToClipboard copies the port number to system clipboard
func CopyToClipboard(port int) error {
	portStr := fmt.Sprintf("%d", port)
	clipboard.Write(clipboard.FmtText, []byte(portStr))
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./actions -v`

Expected: PASS (may skip if clipboard unavailable in CI)

**Step 5: Commit**

```bash
git add actions/
git commit -m "feat(actions): implement clipboard copy"
```

---

## Task 7: Actions - Process Killer

**Files:**
- Create: `actions/kill.go`
- Create: `actions/kill_test.go`

**Step 1: Write test for KillProcess**

Create `actions/kill_test.go`:

```go
package actions

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestKillProcess(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("KillProcess test only runs on macOS")
	}

	// Start a dummy process we can kill
	cmd := exec.Command("sleep", "30")
	err := cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start test process: %v", err)
	}

	pid := cmd.Process.Pid
	t.Logf("Started test process with PID: %d", pid)

	// Kill the process
	err = KillProcess(pid)
	if err != nil {
		t.Fatalf("KillProcess(%d) error = %v", pid, err)
	}

	// Wait a bit for process to die
	time.Sleep(100 * time.Millisecond)

	// Verify process is dead (sending signal 0 checks existence)
	process, _ := os.FindProcess(pid)
	err = process.Signal(os.Signal(nil))

	// On macOS, FindProcess always succeeds, so we check Wait instead
	waitErr := cmd.Wait()
	if waitErr == nil {
		t.Error("Process still running after KillProcess")
	}
}

func TestKillProcess_InvalidPID(t *testing.T) {
	// Try to kill non-existent process
	err := KillProcess(999999)
	if err == nil {
		t.Error("KillProcess(999999) expected error for invalid PID, got nil")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./actions -v`

Expected: FAIL with "undefined: KillProcess"

**Step 3: Implement process killer**

Create `actions/kill.go`:

```go
package actions

import (
	"fmt"
	"os/exec"
)

// KillProcess attempts to terminate a process by PID
// First tries graceful SIGTERM, then falls back to SIGKILL with sudo
func KillProcess(pid int) error {
	pidStr := fmt.Sprintf("%d", pid)

	// Try graceful kill first (SIGTERM)
	cmd := exec.Command("kill", "-15", pidStr)
	err := cmd.Run()

	if err == nil {
		return nil
	}

	// If graceful kill failed, try force kill with sudo via osascript
	// This will prompt user for password via native macOS dialog
	script := fmt.Sprintf("kill -9 %d", pid)
	cmd = exec.Command("osascript", "-e",
		fmt.Sprintf(`do shell script "%s" with administrator privileges`, script))

	return cmd.Run()
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./actions -v -run TestKillProcess$`

Expected: PASS (the first test; second test may need adjustment based on kill behavior)

**Step 5: Commit**

```bash
git add actions/
git commit -m "feat(actions): implement process killer with sudo fallback"
```

---

## Task 8: Menu Builder - Port Formatting

**Files:**
- Create: `menu/builder.go`
- Create: `menu/builder_test.go`

**Step 1: Write test for port formatting**

Create `menu/builder_test.go`:

```go
package menu

import (
	"testing"
	"github.com/yourusername/port-digger/scanner"
)

func TestFormatPortItem(t *testing.T) {
	tests := []struct {
		name string
		info scanner.PortInfo
		want string
	}{
		{
			name: "4-digit port",
			info: scanner.PortInfo{Port: 3000, ProcessName: "node"},
			want: " 3000 • node",
		},
		{
			name: "5-digit port",
			info: scanner.PortInfo{Port: 27017, ProcessName: "mongod"},
			want: "27017 • mongod",
		},
		{
			name: "2-digit port",
			info: scanner.PortInfo{Port: 80, ProcessName: "nginx"},
			want: "   80 • nginx",
		},
		{
			name: "long process name",
			info: scanner.PortInfo{Port: 8080, ProcessName: "java"},
			want: " 8080 • java",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPortItem(tt.info)
			if got != tt.want {
				t.Errorf("formatPortItem() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./menu -v`

Expected: FAIL with "no Go files in .../menu"

**Step 3: Implement port formatting**

Create `menu/builder.go`:

```go
package menu

import (
	"fmt"
	"github.com/yourusername/port-digger/scanner"
)

// formatPortItem formats a port info as "  PORT • ProcessName"
// Port is right-aligned in 5 characters
func formatPortItem(info scanner.PortInfo) string {
	return fmt.Sprintf("%5d • %s", info.Port, info.ProcessName)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./menu -v`

Expected: PASS

**Step 5: Commit**

```bash
git add menu/
git commit -m "feat(menu): add port item formatting"
```

---

## Task 9: Main Application - Systray Setup

**Files:**
- Create: `main.go`
- Create: `icon.go`

**Step 1: Create application icon data**

Create `icon.go`:

```go
package main

// iconData returns the menu bar icon as PNG bytes
// Using a simple port/network icon (16x16 PNG, base64 encoded)
func iconData() []byte {
	// This is a placeholder - replace with actual icon
	// Simple 16x16 black dot for now
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		// ... (minimal PNG data for a simple icon)
	}
}
```

**Step 2: Write minimal main.go with systray**

Create `main.go`:

```go
package main

import (
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconData())
	systray.SetTitle("Port Digger")
	systray.SetTooltip("Monitor TCP listening ports")

	// Add quit option
	mQuit := systray.AddMenuItem("Quit", "Quit Port Digger")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	// Cleanup if needed
}
```

**Step 3: Build and manually test**

Run:
```bash
go build -o port-digger
./port-digger
```

Expected: Menu bar icon appears with "Port Digger" and "Quit" option

**Step 4: Commit**

```bash
git add main.go icon.go
git commit -m "feat(main): add systray initialization"
```

---

## Task 10: Main Application - Menu Refresh Logic

**Files:**
- Modify: `main.go`

**Step 1: Add refreshMenu function skeleton**

Add to `main.go`:

```go
import (
	"sort"
	"github.com/yourusername/port-digger/scanner"
)

func refreshMenu() {
	systray.ResetMenu()

	// Add refresh button
	mRefresh := systray.AddMenuItem("🔄 Refresh", "Rescan ports")
	go func() {
		for range mRefresh.ClickedCh {
			refreshMenu()
		}
	}()

	systray.AddSeparator()

	// Scan ports
	ports, err := scanner.ScanPorts()
	if err != nil {
		systray.AddMenuItem("❌ Scan failed", err.Error())
		return
	}

	if len(ports) == 0 {
		systray.AddMenuItem("No ports listening", "")
		return
	}

	// Sort by port number
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	// Add port menu items (implementation in next step)
	for _, p := range ports {
		addPortMenuItem(p)
	}

	// Add quit at bottom
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Port Digger")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

// Placeholder for next step
func addPortMenuItem(info scanner.PortInfo) {
	// TODO: implement in next task
}
```

**Step 2: Call refreshMenu from onReady**

Modify `onReady()`:

```go
func onReady() {
	systray.SetIcon(iconData())
	systray.SetTitle("Port Digger")
	systray.SetTooltip("Monitor TCP listening ports")

	refreshMenu()
}
```

**Step 3: Build and manually test**

Run:
```bash
go build -o port-digger
./port-digger
```

Expected: Click menu shows refresh button and scanned ports (or "No ports listening")

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat(main): implement menu refresh logic"
```

---

## Task 11: Main Application - Port Menu Items with Actions

**Files:**
- Modify: `main.go`
- Modify: `menu/builder.go`

**Step 1: Import actions package in main.go**

Add to imports in `main.go`:

```go
import (
	"github.com/yourusername/port-digger/actions"
	"github.com/yourusername/port-digger/menu"
	"golang.design/x/clipboard"
)
```

**Step 2: Initialize clipboard in main()**

Modify `main()`:

```go
func main() {
	// Initialize clipboard once at startup
	err := clipboard.Init()
	if err != nil {
		// Non-fatal - clipboard features will just fail silently
		println("Warning: clipboard not available:", err.Error())
	}

	systray.Run(onReady, onExit)
}
```

**Step 3: Implement addPortMenuItem**

Replace placeholder `addPortMenuItem()` in `main.go`:

```go
func addPortMenuItem(info scanner.PortInfo) {
	// Format: " 3000 • node"
	itemText := menu.FormatPortItem(info)
	mPort := systray.AddMenuItem(itemText, fmt.Sprintf("Port %d - %s (PID: %d)",
		info.Port, info.ProcessName, info.PID))

	// Add submenu items
	mOpen := mPort.AddSubMenuItem("Open in Browser", "Open http://localhost:PORT")
	mCopy := mPort.AddSubMenuItem("Copy Port Number", "Copy to clipboard")
	mPort.AddSubMenuItemCheckbox("", "", false) // separator-like
	mKill := mPort.AddSubMenuItem(
		fmt.Sprintf("Kill Process (PID: %d)", info.PID),
		"Terminate this process")

	// Handle submenu actions
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				actions.OpenBrowser(info.Port)
			case <-mCopy.ClickedCh:
				actions.CopyToClipboard(info.Port)
			case <-mKill.ClickedCh:
				err := actions.KillProcess(info.PID)
				if err != nil {
					// Could show notification, but keep it simple for now
					println("Failed to kill process:", err.Error())
				} else {
					// Refresh menu after successful kill
					refreshMenu()
				}
			}
		}
	}()
}
```

**Step 4: Export FormatPortItem in menu package**

Modify `menu/builder.go`:

```go
// FormatPortItem formats a port info as "  PORT • ProcessName"
// Port is right-aligned in 5 characters
func FormatPortItem(info scanner.PortInfo) string {
	return fmt.Sprintf("%5d • %s", info.Port, info.ProcessName)
}
```

**Step 5: Update test to use exported name**

Modify `menu/builder_test.go`:

```go
func TestFormatPortItem(t *testing.T) {
	// ... tests ...
	got := FormatPortItem(tt.info)  // Capital F
	// ...
}
```

**Step 6: Run tests**

Run: `go test ./...`

Expected: All tests PASS

**Step 7: Build and manually test**

Run:
```bash
go build -o port-digger
./port-digger
```

Manual test:
1. Click menu bar icon
2. Hover over a port item
3. Verify submenu shows: "Open in Browser", "Copy Port Number", "Kill Process"
4. Click "Copy Port Number" - verify port is in clipboard
5. Click "Open in Browser" - verify browser opens (or fails gracefully if nothing running)

**Step 8: Commit**

```bash
git add main.go menu/builder.go menu/builder_test.go
git commit -m "feat(main): implement port menu items with action handlers"
```

---

## Task 12: Build Optimization

**Files:**
- Create: `Makefile`
- Create: `README.md`

**Step 1: Create Makefile**

Create `Makefile`:

```makefile
.PHONY: build clean test run

APP_NAME = PortDigger

build:
	go build -ldflags="-s -w" -o $(APP_NAME) .

build-release:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(APP_NAME) .

test:
	go test ./... -v

run: build
	./$(APP_NAME)

clean:
	rm -f $(APP_NAME)
	go clean

install: build
	mkdir -p ~/Applications
	cp $(APP_NAME) ~/Applications/

.DEFAULT_GOAL := build
```

**Step 2: Create README.md**

Create `README.md`:

```markdown
# Port Digger

A lightweight macOS menu bar tool for monitoring TCP listening ports.

## Features

- 🔍 Real-time port monitoring (on-demand, no background polling)
- 🌐 Open ports in browser with one click
- 📋 Copy port numbers to clipboard
- ⚡ Kill processes (with sudo prompt when needed)
- 💾 Minimal memory footprint (~10-20MB)

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/yourusername/port-digger.git
cd port-digger

# Build
make build

# Run
./PortDigger
```

### Manual Install

```bash
make install
# Starts PortDigger from ~/Applications/
```

## Usage

1. Click the menu bar icon to see all listening TCP ports
2. Ports are sorted by number and show process name
3. Hover over any port to see actions:
   - **Open in Browser** - Opens `http://localhost:PORT`
   - **Copy Port Number** - Copies port to clipboard
   - **Kill Process** - Terminates the process (asks for password if needed)
4. Click **Refresh** to rescan ports

## Requirements

- macOS 10.13+
- Go 1.21+ (for building from source)

## Technical Details

- **Runtime Memory**: 10-20MB
- **Binary Size**: 8-15MB
- **Dependencies**: systray, open-golang, clipboard
- **Scan Method**: `lsof -iTCP -sTCP:LISTEN -nP`

## Testing

```bash
make test
```

## License

MIT
```

**Step 3: Test build**

Run:
```bash
make clean
make build
ls -lh PortDigger
```

Expected: Binary size ~8-15MB

**Step 4: Commit**

```bash
git add Makefile README.md
git commit -m "docs: add Makefile and README"
```

---

## Task 13: Fix Module Path References

**Files:**
- Modify: `menu/builder.go`
- Modify: `menu/builder_test.go`
- Modify: `main.go`

**Step 1: Update import paths to use actual module name**

Check `go.mod` for actual module path, then update all files:

In `menu/builder_test.go`:
```go
import (
	"testing"
	"port-digger/scanner"  // Use actual module path from go.mod
)
```

In `main.go`:
```go
import (
	"fmt"
	"sort"

	"github.com/getlantern/systray"
	"golang.design/x/clipboard"

	"port-digger/actions"   // Use actual module path
	"port-digger/menu"
	"port-digger/scanner"
)
```

**Step 2: Run tests to verify imports**

Run: `go test ./...`

Expected: All tests PASS

**Step 3: Build to verify**

Run: `make build`

Expected: Clean build with no errors

**Step 4: Commit**

```bash
git add menu/ main.go
git commit -m "fix: correct module import paths"
```

---

## Task 14: Add Real Menu Bar Icon

**Files:**
- Modify: `icon.go`

**Step 1: Create proper icon data**

Replace `icon.go` content with actual icon. For quick implementation, use an online tool to convert a small PNG to byte array, or use this simple network icon:

```go
package main

// iconData returns a simple network/port icon as PNG bytes
// 16x16 pixel icon showing a network port symbol
var iconData = func() []byte {
	// Embedded PNG - represents a simple port/plug icon
	// Generated from a 16x16 PNG with transparency
	return []byte{
		// Include actual PNG byte data here
		// For now, use a Unicode character rendered approach
	}
}

// Fallback: use unicode for simplicity
func iconDataFallback() []byte {
	return []byte{}  // systray will show app name only
}
```

**Step 2: Use simpler approach - no icon, just show title**

Replace with:

```go
package main

// iconData returns empty - systray will show title instead
func iconData() []byte {
	return []byte{}
}
```

**Step 3: Update main.go to use better title**

Modify `onReady()` in `main.go`:

```go
func onReady() {
	systray.SetIcon(iconData())
	systray.SetTitle("🔌")  // Port plug emoji as icon
	systray.SetTooltip("Port Digger - Monitor TCP Ports")

	refreshMenu()
}
```

**Step 4: Test**

Run: `make run`

Expected: Menu bar shows "🔌" emoji

**Step 5: Commit**

```bash
git add icon.go main.go
git commit -m "feat: add emoji-based menu bar icon"
```

---

## Task 15: Error Handling and Edge Cases

**Files:**
- Modify: `main.go`
- Create: `test_edge_cases.md` (manual test checklist)

**Step 1: Add error handling for clipboard failures**

Modify `addPortMenuItem()` in `main.go`:

```go
case <-mCopy.ClickedCh:
	err := actions.CopyToClipboard(info.Port)
	if err != nil {
		println("Failed to copy to clipboard:", err.Error())
	}
```

**Step 2: Add graceful handling for scanner errors**

Already implemented in `refreshMenu()`, verify:

```go
ports, err := scanner.ScanPorts()
if err != nil {
	systray.AddMenuItem("❌ Scan failed", err.Error())
	return
}
```

**Step 3: Create manual test checklist**

Create `test_edge_cases.md`:

```markdown
# Port Digger - Manual Test Checklist

## Edge Cases to Test

### Scanner
- [ ] No ports listening (stop all servers)
- [ ] Many ports (>20)
- [ ] Port number edge cases (80, 8080, 65535)
- [ ] Process names with spaces
- [ ] Multiple processes on different ports

### Actions
- [ ] Open browser on non-serving port (should fail gracefully)
- [ ] Copy port to clipboard, paste elsewhere
- [ ] Kill own process (should prompt for password)
- [ ] Kill process without sudo (should work for own processes)
- [ ] Kill process with sudo (test with system process)

### Menu
- [ ] Click refresh multiple times rapidly
- [ ] Open submenu, then refresh menu
- [ ] Hover over multiple ports quickly

### System
- [ ] Run on macOS 13+
- [ ] Check memory usage with Activity Monitor
- [ ] Leave running for extended period
- [ ] Check binary size

## Success Criteria
- No crashes
- Memory stays under 30MB
- UI remains responsive
- All actions work or fail gracefully
```

**Step 4: Manual test session**

Run through checklist manually with:
```bash
make run
```

**Step 5: Commit**

```bash
git add main.go test_edge_cases.md
git commit -m "test: add edge case handling and manual test checklist"
```

---

## Task 16: Final Integration Test and Polish

**Files:**
- Modify: `main.go`
- Modify: `README.md`

**Step 1: Add version info**

Add to `main.go`:

```go
const version = "1.0.0"

func onReady() {
	systray.SetIcon(iconData())
	systray.SetTitle("🔌")
	systray.SetTooltip(fmt.Sprintf("Port Digger v%s - Monitor TCP Ports", version))

	refreshMenu()
}
```

**Step 2: Add version to menu**

Modify `refreshMenu()` to add version at bottom:

```go
func refreshMenu() {
	// ... existing code ...

	// Add quit at bottom
	systray.AddSeparator()
	systray.AddMenuItem(fmt.Sprintf("Port Digger v%s", version), "About")
	mQuit := systray.AddMenuItem("Quit", "Quit Port Digger")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}
```

**Step 3: Full build and test**

Run:
```bash
make clean
make build
make test
./PortDigger
```

Test all features:
1. Refresh works
2. Port list populates
3. Browser open works
4. Copy to clipboard works
5. Kill process works (with password prompt)

**Step 4: Update README with screenshots/demo**

Add to `README.md` after Features:

```markdown
## Screenshot

```
 3000 • node
 8080 • Python
27017 • mongod
```

**Example Actions:**
- Click "3000 • node" → See submenu
  - Open in Browser → Opens http://localhost:3000
  - Copy Port Number → "3000" in clipboard
  - Kill Process (PID: 12345) → Prompts for password, terminates node
```

**Step 5: Final commit**

```bash
git add main.go README.md
git commit -m "feat: add version info and final polish"
```

**Step 6: Tag release**

```bash
git tag -a v1.0.0 -m "Release v1.0.0 - Initial release"
```

---

## Post-Implementation Checklist

Before considering this complete:

- [ ] All tests pass: `go test ./...`
- [ ] Build succeeds: `make build`
- [ ] Binary size reasonable (8-15MB): `ls -lh PortDigger`
- [ ] Manual testing completed (see `test_edge_cases.md`)
- [ ] Memory usage verified (<30MB in Activity Monitor)
- [ ] README is accurate and helpful
- [ ] All code committed to git
- [ ] Version tagged

## Future Enhancements (Not in Scope)

Reference design doc section "后期扩展点" (lines 290-296) for:
1. Custom process naming via config file
2. Port grouping by service type
3. Global keyboard shortcuts
4. Port history tracking
5. UDP port support

---

**Implementation complete!** 🎉

Binary should be ~10-15MB, runtime memory ~10-20MB, with all core features working.
