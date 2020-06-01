package echomiddleware

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// 流用元：https://github.com/labstack/echo/blob/master/middleware/recover.go

type (
	// RecoverConfig defines the config for Recover middleware.
	RecoverConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Size of the stack to be printed.
		// Optional. Default value 4KB.
		StackSize int `yaml:"stack_size"`

		// DisableStackAll disables formatting stack traces of all other goroutines
		// into buffer after the trace for the current goroutine.
		// Optional. Default value false.
		DisableStackAll bool `yaml:"disable_stack_all"`

		// DisablePrintStack disables printing stack trace.
		// Optional. Default value as false.
		DisablePrintStack bool `yaml:"disable_print_stack"`
	}
)

var (
	// DefaultRecoverConfig is the default Recover middleware config.
	DefaultRecoverConfig = RecoverConfig{
		Skipper:           middleware.DefaultSkipper,
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   false,
		DisablePrintStack: false,
	}
)

// GCPRecover returns a middleware which recovers from panics anywhere in the chain
// and handles the control to the centralized HTTPErrorHandler.
func GCPRecover() echo.MiddlewareFunc {
	return GCPRecoverWithConfig(DefaultRecoverConfig)
}

// GCPRecoverWithConfig returns a Recover middleware with config.
// See: `Recover()`.
func GCPRecoverWithConfig(config RecoverConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultRecoverConfig.Skipper
	}
	if config.StackSize == 0 {
		config.StackSize = DefaultRecoverConfig.StackSize
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					if !config.DisablePrintStack {
						// Logrusを用いてログ出力する
						log := logrus.New()
						log.Level = logrus.DebugLevel
						log.Formatter = &logrus.JSONFormatter{
							FieldMap: logrus.FieldMap{
								logrus.FieldKeyTime:  "time",
								logrus.FieldKeyLevel: "severity",
								logrus.FieldKeyMsg:   "message",
							},
							TimestampFormat: time.RFC3339Nano,
						}
						log.Out = os.Stderr
						log.WithFields(logrus.Fields{
							"@type": "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
						}).Errorf("panic(%v): %s", err, stack[:length])
						// @todo: Fatal, Panicの場合exit()が呼ばれるためErrorにしている
						//      : 一度文字列にしてFatalに置換するか、
						//      : logrus.FieldKeyLevel: "dump"とし、WithFields(logrus.Fields{ "severity": "fatal"...
						//      : などで対策したほうがいいかも...
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}
