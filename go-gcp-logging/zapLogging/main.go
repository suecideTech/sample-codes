package main

import (
	"fmt"
	"net/http"
	"runtime"

	"zapLogging/echomw"
	"zapLogging/zaplogger"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	panicStackSize int  = 4 << 10
	panicStackAll  bool = true
)

var (
	logger    *zap.Logger
	errLogger *zap.Logger
)

func main() {
	logger = zaplogger.NewLogger()
	errLogger = zaplogger.NewErrorLogger()

	// Catch Panic.
	defer func() {
		if x := recover(); x != nil {
			stack := make([]byte, panicStackSize)
			length := runtime.Stack(stack, panicStackAll)
			errLogger.Panic(fmt.Sprintf("%v", x), zap.ByteString("stack", stack[:length]))
		}
	}()

	// Setup Echo.
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(echomw.ZapLogger(logger))
	e.Use(echomw.GCPRecoverWithConfig(errLogger,
		echomw.RecoverConfig{
			StackSize:       panicStackSize,
			DisableStackAll: !panicStackAll,
		}))

	e.GET("/", func(c echo.Context) error {

		logger.Info("Hello World!")
		logger.Info("Fields Test.", zap.String("TestFields", "TestValue"))

		return c.String(http.StatusOK, "Hello, World!\n")
	})

	e.GET("/fatal", func(c echo.Context) error {

		logger.Fatal("Hello Fatal!", zap.String("Method", c.Request().Method))

		return c.String(http.StatusOK, "Hello, World!\n")
	})

	e.GET("/panic", func(c echo.Context) error {

		logger.Panic("Hello Panic!", zap.String("Method", c.Request().Method))

		return c.String(http.StatusOK, "Hello, World!\n")
	})

	//nilPointerTesting()

	logger.Info("Echo Initialize Complete! ListenPort(80)")
	logger.Fatal(e.Start(":80").Error())
}

func nilPointerTesting() {
	var p *interface{}
	*p = 0
}
