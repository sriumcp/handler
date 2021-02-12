package common

import (
	"testing"

	"github.com/iter8-tools/handler/base"
	"github.com/stretchr/testify/assert"
)

func TestMakeTask(t *testing.T) {
	task, err := MakeTask(&base.TaskSpec{
		Library: "common",
		Task:    "exec",
		With: map[string]interface{}{
			"cmd": "your-command",
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	task, err = MakeTask(&base.TaskSpec{
		Library: "common",
		Task:    "run",
		With: map[string]interface{}{
			"cmd": "your-command",
		},
	})
	assert.Nil(t, task)
	assert.Error(t, err)
}
