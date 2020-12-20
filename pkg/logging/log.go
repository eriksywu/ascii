package logging

import (
	"flag"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Logger *StandardEventLogger

func init() {
	flag.BoolVar(&logToFile, "logToFile", false, "enable logging to local file ascii.log")
	Logger = NewLogger()
}

const (
	logFilePath = "ascii.log"
)

type StandardEventLogger struct {
	*logrus.Logger
}

var logToFile bool

func NewLogger() *StandardEventLogger {
	logrusLogger := logrus.New()

	if logToFile {
		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			os.Create(logFilePath)
		}
		if logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err == nil {
			muxWriter := io.MultiWriter(os.Stdout, logFile)
			logrusLogger.SetOutput(muxWriter)
		}
	}
	formatter := &runtime.Formatter{
		ChildFormatter: &logrus.JSONFormatter{},
		Line:           true,
		File:           true,
	}
	logger := &StandardEventLogger{
		Logger: logrusLogger,
	}

	logger.Formatter = formatter
	return logger
}
