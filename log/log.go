package log

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// Level refers to the level of logging
type Level string

// Logger is an interface that needs to be implemented in order to log.
type Logger interface {
	Debug(...interface{})
	Debugln(...interface{})
	Debugf(string, ...interface{})

	Info(...interface{})
	Infoln(...interface{})
	Infof(string, ...interface{})

	Warn(...interface{})
	Warnln(...interface{})
	Warnf(string, ...interface{})
	Warningf(string, ...interface{})

	Error(...interface{})
	Errorln(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalln(...interface{})
	Fatalf(string, ...interface{})

	Panic(...interface{})
	Panicln(...interface{})
	Panicf(string, ...interface{})

	With(key string, value interface{}) Logger
	WithField(key string, value interface{}) Logger
	Set(level Level) error
}

type logger struct {
	entry *logrus.Entry
}

func (l logger) Debug(args ...interface{}) {
	l.sourced().Debug(args...)
}

func (l logger) Debugln(args ...interface{}) {
	l.sourced().Debugln(args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.sourced().Debugf(format, args...)
}

func (l logger) Info(args ...interface{}) {
	l.sourced().Info(args...)
}

func (l logger) Infoln(args ...interface{}) {
	l.sourced().Infoln(args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.sourced().Infof(format, args...)
}

func (l logger) Warn(args ...interface{}) {
	l.sourced().Warn(args...)
}

func (l logger) Warnln(args ...interface{}) {
	l.sourced().Warnln(args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.sourced().Warnf(format, args...)
}

func (l logger) Warningf(format string, args ...interface{}) {
	l.sourced().Warnf(format, args...)
}

func (l logger) Error(args ...interface{}) {
	l.sourced().Error(args...)
}

func (l logger) Errorln(args ...interface{}) {
	l.sourced().Errorln(args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.sourced().Errorf(format, args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.sourced().Fatal(args...)
}

func (l logger) Fatalln(args ...interface{}) {
	l.sourced().Fatalln(args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.sourced().Fatalf(format, args...)
}
func (l logger) Panic(args ...interface{}) {
	l.sourced().Panic(args...)
}
func (l logger) Panicln(args ...interface{}) {
	l.sourced().Panicln(args...)
}
func (l logger) Panicf(format string, args ...interface{}) {
	l.sourced().Panicf(format, args...)
}

func (l logger) With(key string, value interface{}) Logger {
	return &logger{l.entry.WithField(key, value)}
}

func (l logger) WithField(key string, value interface{}) Logger {
	return &logger{l.entry.WithField(key, value)}
}

func (l *logger) Set(level Level) error {
	leLev, err := logrus.ParseLevel(string(level))
	if err != nil {
		return err
	}
	l.entry.Logger.Level = leLev
	return nil
}

func (l logger) sourced() *logrus.Entry {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		file = file[slash+1:]
	}
	return l.entry.WithField("src", fmt.Sprintf("%s:%d", file, line))
}

var baseLogger = &logger{
	entry: &logrus.Entry{
		Logger: logrus.New(),
	},
}

// Base returns the base logger
func Base() Logger {
	return baseLogger
}

// Debug logs debug message
func Debug(args ...interface{}) {
	baseLogger.sourced().Debug(args...)
}

// Debugln logs debug message
func Debugln(args ...interface{}) {
	baseLogger.sourced().Debugln(args...)
}

// Debugf logs debug message
func Debugf(format string, args ...interface{}) {
	baseLogger.sourced().Debugf(format, args...)
}

// Info logs info message
func Info(args ...interface{}) {
	baseLogger.sourced().Info(args...)
}

// Infoln logs info message
func Infoln(args ...interface{}) {
	baseLogger.sourced().Infoln(args...)
}

// Infof logs info message
func Infof(format string, args ...interface{}) {
	baseLogger.sourced().Infof(format, args...)
}

// Warn logs warn message
func Warn(args ...interface{}) {
	baseLogger.sourced().Warn(args...)
}

// Warnln logs warn message
func Warnln(args ...interface{}) {
	baseLogger.sourced().Warnln(args...)
}

// Warnf logs warn message
func Warnf(format string, args ...interface{}) {
	baseLogger.sourced().Warnf(format, args...)
}

// Error logs error message
func Error(args ...interface{}) {
	baseLogger.sourced().Error(args...)
}

// Errorln logs error message
func Errorln(args ...interface{}) {
	baseLogger.sourced().Errorln(args...)
}

// Errorf logs error message
func Errorf(format string, args ...interface{}) {
	baseLogger.sourced().Errorf(format, args...)
}

// Fatal logs fatal message
func Fatal(args ...interface{}) {
	baseLogger.sourced().Fatal(args...)
}

// Fatalln logs fatal message
func Fatalln(args ...interface{}) {
	baseLogger.sourced().Fatalln(args...)
}

// Fatalf logs fatal message
func Fatalf(format string, args ...interface{}) {
	baseLogger.sourced().Fatalf(format, args...)
}

// With adds a key:value to the logger
func With(key string, value interface{}) Logger {
	return baseLogger.With(key, value)
}

// WithField adds a key:value to the logger
func WithField(key string, value interface{}) Logger {
	return baseLogger.WithField(key, value)
}

// Set will set the logger level
func Set(level Level) error {
	return baseLogger.Set(level)
}

// Panic logs panic message
func Panic(args ...interface{}) {
	baseLogger.Panic(args...)
}

// Panicln logs panicln message
func Panicln(args ...interface{}) {
	baseLogger.Panicln(args...)
}

// Panicf logs panicln message
func Panicf(format string, args ...interface{}) {
	baseLogger.Panicf(format, args...)
}
