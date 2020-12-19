package logging

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Logger *StandardEventLogger

func init() {
	Logger = NewLogger()
}

const (
	logFilePath = "ascii.log"
)

type StandardEventLogger struct {
	*logrus.Logger
}

func NewLogger() *StandardEventLogger {
	logrusLogger := logrus.New()
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		os.Create(logFilePath)
	}
	if logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err == nil {
		muxWriter := io.MultiWriter(os.Stdout, logFile)
		logrusLogger.SetOutput(muxWriter)
	}
	logrusLogger.Formatter = &logrus.JSONFormatter{}
	logger := &StandardEventLogger{
		Logger: logrusLogger,
	}
	return logger
}
