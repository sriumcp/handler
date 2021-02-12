package knative

import (
	"testing"

	"github.com/iter8-tools/handler/base"
	"github.com/stretchr/testify/assert"
)

func TestMakeTask(t *testing.T) {
	task, err := MakeTask(&base.TaskSpec{
		Library: "knative",
		Task:    "init-experiment",
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	task, err = MakeTask(&base.TaskSpec{
		Library: "knative",
		Task:    "init-experimental",
	})
	assert.Nil(t, task)
	assert.Error(t, err)
}
