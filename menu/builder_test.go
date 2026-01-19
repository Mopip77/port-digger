package menu

import (
	"port-digger/scanner"
	"testing"
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
			got := FormatPortItem(tt.info)  // Capital F
			if got != tt.want {
				t.Errorf("FormatPortItem() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatPortItemWithRewrite(t *testing.T) {
	tests := []struct {
		name        string
		info        scanner.PortInfo
		rewriteName string
		want        string
	}{
		{
			name:        "with rewritten name",
			info:        scanner.PortInfo{Port: 3000, ProcessName: "node"},
			rewriteName: "claude-code-ui",
			want:        " 3000 • node (claude-code-ui✨)",
		},
		{
			name:        "empty rewrite falls back",
			info:        scanner.PortInfo{Port: 3000, ProcessName: "node"},
			rewriteName: "",
			want:        " 3000 • node",
		},
		{
			name:        "未知 rewrite falls back",
			info:        scanner.PortInfo{Port: 8080, ProcessName: "python"},
			rewriteName: "未知",
			want:        " 8080 • python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPortItemWithRewrite(tt.info, tt.rewriteName)
			if got != tt.want {
				t.Errorf("FormatPortItemWithRewrite() = %q, want %q", got, tt.want)
			}
		})
	}
}
