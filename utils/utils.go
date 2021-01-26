// Package utils supplies logging and testing utils.
package utils

import (
	iter8utils "github.com/iter8-tools/etc3/util"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

var logLevel logrus.Level = logrus.InfoLevel

// SetLogLevel sets level for logging.
func SetLogLevel(l logrus.Level) {
	logLevel = l
	if log != nil {
		log.SetLevel(logLevel)
	}
}

// GetLogger returns a logger, if needed after creating it.
func GetLogger() *logrus.Logger {
	if log == nil {
		log = logrus.New()
		log.SetLevel(logLevel)
	}
	return log
}

// CompletePath determines complete path of a file
var CompletePath func(prefix string, suffix string) string = iter8utils.CompletePath
