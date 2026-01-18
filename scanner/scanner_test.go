package scanner

import (
	"runtime"
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
