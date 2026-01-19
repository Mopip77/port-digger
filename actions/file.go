package actions

import (
	"github.com/skratchdot/open-golang/open"
)

// OpenFile opens a file in the default application
func OpenFile(path string) error {
	return open.Start(path)
}
