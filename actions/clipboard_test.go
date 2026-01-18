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
