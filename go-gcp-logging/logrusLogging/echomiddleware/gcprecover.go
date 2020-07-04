package echomiddleware

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

type logEntryWithErrorReporting struct {
	Timestamp string `json:"time"`
	Severity  string `json:"severity"`
	Type      string `json:"@type"`
	Message   string `json:"message"`
}

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
						message := fmt.Sprintf("panic(%v): %s", err, stack[:length])
						now := time.Now()
						logContents := logEntryWithErrorReporting{
							Timestamp: now.Format(time.RFC3339Nano),
							Severity:  "fatal", // Error Reportingへ挿入するseverityはfatal固定にする
							Type:      "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent",
							Message:   message,
						}
						logMessage, err := json.Marshal(logContents)
						if err != nil {
							return
						}
						log.SetOutput(os.Stdout)
						log.Print(string(logMessage))
					}
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}
