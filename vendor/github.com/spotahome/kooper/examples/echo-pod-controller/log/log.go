package log

import (
	"github.com/spotahome/kooper/log"
)

// Logger is the interface of the controller logger. This is an example
// so our Loggger will be the same as the kooper one.
type Logger interface {
	log.Logger
}
