package logging

import (
	"github.com/sirupsen/logrus"
)

var Logger *StandardEventLogger

func init() {
	Logger = NewLogger()
}

type StandardEventLogger struct {
	*logrus.Logger
}

func NewLogger() *StandardEventLogger {
	logrusLogger := logrus.New()
	logrusLogger.Formatter = &logrus.JSONFormatter{}
	logger := &StandardEventLogger{
		Logger: logrusLogger,
	}
	return logger
}
