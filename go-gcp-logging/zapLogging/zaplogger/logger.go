package zaplogger

import (
	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
)

// Logger Zapdriver Logger
var Logger *zap.Logger

// ErrorLogger Zapdriver Logger(Error)
var ErrorLogger *zap.Logger

// NewLogger returns logger
func NewLogger() *zap.Logger {
	Logger, err := zapdriver.NewProduction()
	if err != nil {
		return nil
	}
	return Logger
}

// NewErrorLogger returns logger(error)
func NewErrorLogger() *zap.Logger {
	ErrorLogger, err := zapdriver.NewProductionWithCore(zapdriver.WrapCore(
		zapdriver.ReportAllErrors(true),
	))
	if err != nil {
		return nil
	}
	return ErrorLogger
}
