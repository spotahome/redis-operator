package log

import (
	"fmt"
	"log"
)

// KV is a helper type for structured logging fields usage.
type KV map[string]interface{}

// Logger is the interface that the loggers used by the library will use.
type Logger interface {
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	WithKV(KV) Logger
}

// Dummy logger doesn't log anything.
const Dummy = dummy(0)

type dummy int

func (d dummy) Infof(format string, args ...interface{})    {}
func (d dummy) Warningf(format string, args ...interface{}) {}
func (d dummy) Errorf(format string, args ...interface{})   {}
func (d dummy) Debugf(format string, args ...interface{})   {}
func (d dummy) WithKV(KV) Logger                            { return d }

// Std is a wrapper for go standard library logger.
type std struct {
	debug  bool
	fields map[string]interface{}
}

// NewStd returns a Logger implementation with the standard logger.
func NewStd(debug bool) Logger {
	return std{
		debug:  debug,
		fields: map[string]interface{}{},
	}
}

func (s std) logWithPrefix(prefix, format string, kv map[string]interface{}, args ...interface{}) {

	msgFmt := ""
	if len(kv) == 0 {
		msgFmt = fmt.Sprintf("%s\t%s", prefix, format)
	} else {
		msgFmt = fmt.Sprintf("%s\t%s\t\t%v", prefix, format, kv)
	}

	log.Printf(msgFmt, args...)
}

func (s std) Infof(format string, args ...interface{}) {
	s.logWithPrefix("[INFO]", format, s.fields, args...)
}
func (s std) Warningf(format string, args ...interface{}) {
	s.logWithPrefix("[WARN]", format, s.fields, args...)
}
func (s std) Errorf(format string, args ...interface{}) {
	s.logWithPrefix("[ERROR]", format, s.fields, args...)
}
func (s std) Debugf(format string, args ...interface{}) {
	if s.debug {
		s.logWithPrefix("[DEBUG]", format, s.fields, args...)
	}
}

func (s std) WithKV(kv KV) Logger {
	kvs := map[string]interface{}{}
	for k, v := range s.fields {
		kvs[k] = v
	}
	for k, v := range kv {
		kvs[k] = v
	}

	return std{debug: s.debug, fields: kvs}
}
