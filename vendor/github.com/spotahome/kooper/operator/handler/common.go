package handler

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/spotahome/kooper/log"
)

// Logger will log the handling events. This handler can be sued to test that a
//  controller receives resource events..
type Logger struct {
	logger log.Logger
}

// NewLogger returns a new logger.
func NewLogger(logger log.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

// Add satisfies Handler interface.
func (l *Logger) Add(obj runtime.Object) error {
	l.logger.Infof("event add: %#v", obj)
	return nil
}

// Delete satisfies Handler interface.
func (l *Logger) Delete(objKey string) error {
	l.logger.Infof("event delete: %#v", objKey)
	return nil
}
