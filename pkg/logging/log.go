package logging

import (
	"context"
	"github.com/sirupsen/logrus"
)

var Logger *StandardEventLogger

func init() {
	Logger = NewLogger()
}

const (
	CorrelationID = "correlationID"
	OperationName = "OperationName"
)

type StandardEventLogger struct {
	*logrus.Logger
}

func GetLogEntry(context context.Context, logger *StandardEventLogger) *logrus.Entry {
	entry := logrus.NewEntry(logger.Logger)
	correlationID := context.Value(CorrelationID)
	if correlationID != nil {
		entry.WithField(CorrelationID, correlationID)
	}
	operationName := context.Value(OperationName)
	if operationName != nil {
		entry.WithField(OperationName, operationName)
	}
	return entry
}

func NewLogger() *StandardEventLogger {
	logrusLogger := logrus.New()
	logrusLogger.Formatter = &logrus.JSONFormatter{}
	logger := &StandardEventLogger{
		Logger: logrusLogger,
	}
	return logger
}
