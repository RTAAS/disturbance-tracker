// +build headless

package review

import (
	"dtrack/log"
)

// Primary post-bootstrap entry point
func Launch() {
	log.Die("This version was built using headless mode (GUI not included)")
}
