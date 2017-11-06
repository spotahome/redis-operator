package log

// Nil is a nil logger
var Nil = nilLogger{}

type nilLogger struct{}

func (n nilLogger) Debug(...interface{})                           {}
func (n nilLogger) Debugln(...interface{})                         {}
func (n nilLogger) Debugf(string, ...interface{})                  {}
func (n nilLogger) Info(...interface{})                            {}
func (n nilLogger) Infoln(...interface{})                          {}
func (n nilLogger) Infof(string, ...interface{})                   {}
func (n nilLogger) Warn(...interface{})                            {}
func (n nilLogger) Warnln(...interface{})                          {}
func (n nilLogger) Warnf(string, ...interface{})                   {}
func (n nilLogger) Error(...interface{})                           {}
func (n nilLogger) Errorln(...interface{})                         {}
func (n nilLogger) Errorf(string, ...interface{})                  {}
func (n nilLogger) Fatal(...interface{})                           {}
func (n nilLogger) Fatalln(...interface{})                         {}
func (n nilLogger) Fatalf(string, ...interface{})                  {}
func (n nilLogger) Panic(...interface{})                           {}
func (n nilLogger) Panicln(...interface{})                         {}
func (n nilLogger) Panicf(string, ...interface{})                  {}
func (n nilLogger) With(key string, value interface{}) Logger      { return n }
func (n nilLogger) WithField(key string, value interface{}) Logger { return n }
func (n nilLogger) Set(level Level) error                          { return nil }
