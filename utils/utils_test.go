package utils

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMakeTask(t *testing.T) {
	log := GetLogger()
	log.Info("hello world")
	assert.NotEmpty(t, log)
	SetLogLevel(logrus.InfoLevel)
}
