package log

// Dummy is a dummy logger
var Dummy = DummyLogger{}

// DummyLogger is an empty logger mainly used for tests
type DummyLogger struct{}

func (l DummyLogger) Debug(...interface{})                           {}
func (l DummyLogger) Debugln(...interface{})                         {}
func (l DummyLogger) Debugf(string, ...interface{})                  {}
func (l DummyLogger) Info(...interface{})                            {}
func (l DummyLogger) Infoln(...interface{})                          {}
func (l DummyLogger) Infof(string, ...interface{})                   {}
func (l DummyLogger) Warn(...interface{})                            {}
func (l DummyLogger) Warnln(...interface{})                          {}
func (l DummyLogger) Warnf(string, ...interface{})                   {}
func (l DummyLogger) Warningf(format string, args ...interface{})    {}
func (l DummyLogger) Error(...interface{})                           {}
func (l DummyLogger) Errorln(...interface{})                         {}
func (l DummyLogger) Errorf(string, ...interface{})                  {}
func (l DummyLogger) Fatal(...interface{})                           {}
func (l DummyLogger) Fatalln(...interface{})                         {}
func (l DummyLogger) Fatalf(string, ...interface{})                  {}
func (l DummyLogger) Panic(...interface{})                           {}
func (l DummyLogger) Panicln(...interface{})                         {}
func (l DummyLogger) Panicf(string, ...interface{})                  {}
func (l DummyLogger) With(key string, value interface{}) Logger      { return l }
func (l DummyLogger) WithField(key string, value interface{}) Logger { return l }
func (l DummyLogger) Set(level Level) error                          { return nil }
