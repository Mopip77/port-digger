package scanner

// PortInfo represents a listening TCP port with associated process information
type PortInfo struct {
	Port        int    // Port number
	ProcessName string // Process name from lsof COMMAND column
	PID         int    // Process ID
	Command     string // Full command line (for future custom naming)
	Protocol    string // "TCP" or "TCP6"
}
