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
