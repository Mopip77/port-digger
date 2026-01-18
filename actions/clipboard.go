package actions

import (
	"fmt"
	"golang.design/x/clipboard"
)

// CopyToClipboard copies the port number to system clipboard
func CopyToClipboard(port int) error {
	portStr := fmt.Sprintf("%d", port)
	clipboard.Write(clipboard.FmtText, []byte(portStr))
	return nil
}
