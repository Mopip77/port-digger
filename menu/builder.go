package menu

import (
	"fmt"
	"port-digger/scanner"
)

// FormatPortItem formats a port info as "  PORT • ProcessName"
// Port is right-aligned in 5 characters
func FormatPortItem(info scanner.PortInfo) string {
	return fmt.Sprintf("%5d • %s", info.Port, info.ProcessName)
}
