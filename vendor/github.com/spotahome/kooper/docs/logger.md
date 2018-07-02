# Logger

In kooper everything is pluggable, this is because it embraces dependency injection in all of its components, that's why you can plug your own logger, the only requirement is to implement the [`Logger`][logger-interface] interface.

```golang
type Logger interface {
    Infof(format string, args ...interface{})
    Warningf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
}
```

Although it comes with some loggers by default.

## Available loggers

If you don't need a custom logger, you can pass

* Dummy: doesn't log anything (mainly for the tests).
* Glog: Uses [Glog][glog] logger (this logger is a global logger)
* Std: Uses default go logger (this logger is a global logger)

## Use a logger in the controller

```golang
import (
    "github.com/spotahome/kooper/log"
    ...
)
...

log := &log.Glog{}

...

ctrl := controller.NewSequential(30*time.Second, hand, retr, m, log)
...
```

**Note: if you pass nil as the logger to the controller it will use `log.Std` logger by default**


[logger-interface]: https://github.com/spotahome/kooper/blob/master/log/log.go
[glog]: https://github.com/golang/glog