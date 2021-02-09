package common

import (
	"errors"

	"github.com/iter8-tools/handler/base"
	"github.com/iter8-tools/handler/utils"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = utils.GetLogger()
}

// MakeTask constructs a Task from a TaskSpec or returns an error if any.
func MakeTask(t *base.TaskSpec) (base.Task, error) {
	switch t.Task {
	case "exec":
		return MakeExec(t)
	default:
		return nil, errors.New("Unknown task: " + t.Task)
	}
}
