package menu

import (
	"testing"
	"port-digger/scanner"
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
