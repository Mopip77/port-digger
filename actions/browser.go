package actions

import (
	"fmt"
	"github.com/skratchdot/open-golang/open"
)

// formatURL creates the localhost URL for a given port
func formatURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

// OpenBrowser opens the default browser to the given port
func OpenBrowser(port int) error {
	url := formatURL(port)
	return open.Run(url)
}
