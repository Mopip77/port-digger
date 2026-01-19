package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PortInfo represents a listening TCP port with associated process information
type PortInfo struct {
	Port        int    // Port number
	ProcessName string // Process name from lsof COMMAND column
	PID         int    // Process ID
	Command     string // Full command line (for future custom naming)
	Protocol    string // "TCP" or "TCP6"
}

// GetFullCommand retrieves the full command line for a process by PID
// Uses: ps -p <pid> -o command=
func GetFullCommand(pid int) string {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return ""
	}

	return strings.TrimSpace(stdout.String())
}

// unescapeLsofString decodes hex escape sequences like \x20 to actual characters
// lsof escapes special characters in command names using \xHH notation
func unescapeLsofString(s string) string {
	if !strings.Contains(s, "\\x") {
		return s
	}

	var result strings.Builder
	i := 0
	for i < len(s) {
		if i+3 < len(s) && s[i] == '\\' && s[i+1] == 'x' {
			// Parse the two hex digits
			hexStr := s[i+2 : i+4]
			if val, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
				result.WriteByte(byte(val))
				i += 4
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

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
	// Actually, field 8 is just the protocol like "TCP" and the address is in later fields
	// The full line after field 7 looks like: "TCP *:3000 (LISTEN)"
	// Let's reconstruct: find the protocol in field 7, and address in field 8

	// Field 7 is TYPE, field 8 is NAME
	// Field 7 should be "TCP" or "TCP6" (looking at actual lsof output)
	// Actually on further review: the format is more complex
	// Let me use a simpler approach - find "TCP" or "TCP6" followed by address

	protocol := ""
	portStr := ""

	// Look for TCP or TCP6 in the fields
	for i := 7; i < len(fields); i++ {
		if fields[i] == "TCP" || fields[i] == "TCP6" {
			protocol = fields[i]
			// Next field should have the address
			if i+1 < len(fields) {
				portStr = fields[i+1]
			}
			break
		}
	}

	if protocol == "" || portStr == "" {
		return nil, fmt.Errorf("could not find protocol and port")
	}

	// Extract port from "*:3000" or "127.0.0.1:8080"
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
		ProcessName: unescapeLsofString(fields[0]),
		PID:         pid,
		Command:     unescapeLsofString(fields[0]),
		Protocol:    protocol,
	}, nil
}

// ScanPorts executes lsof to get all listening TCP ports
func ScanPorts() ([]PortInfo, error) {
	// Execute: lsof +c 0 -iTCP -sTCP:LISTEN -nP
	// +c 0 shows full command name without truncation
	cmd := exec.Command("lsof", "+c", "0", "-iTCP", "-sTCP:LISTEN", "-nP")

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
