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

// FormatPortItemWithRewrite formats a port info with a rewritten service name
// Format: "  PORT • ProcessName (ServiceName)"
// If rewrittenName is empty or "未知", falls back to FormatPortItem
func FormatPortItemWithRewrite(info scanner.PortInfo, rewrittenName string) string {
	if rewrittenName == "" || rewrittenName == "未知" || rewrittenName == info.ProcessName {
		return FormatPortItem(info)
	}
	return fmt.Sprintf("%5d • %s (%s✨)", info.Port, info.ProcessName, rewrittenName)
}
