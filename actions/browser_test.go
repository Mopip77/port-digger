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
