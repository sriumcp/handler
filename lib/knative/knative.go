package knative

import (
	"errors"

	"github.com/iter8-tools/handler/base"
)

// MakeTask constructs a Task from a TaskMeta or returns an error if any.
func MakeTask(t *base.TaskMeta) (base.Task, error) {
	switch t.Task {
	default:
		return nil, errors.New("Unknown task: " + t.Task)
	}
}
