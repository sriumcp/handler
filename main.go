// Package main is the entry point for handler
package main

import (
	"github.com/iter8-tools/handler/cmd"
	"github.com/iter8-tools/handler/utils"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = utils.GetLogger()
}

func main() {
	cmd.Execute()
}
