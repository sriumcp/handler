package common

import (
	"encoding/json"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeTask(t *testing.T) {
	b, _ := json.Marshal("your-command")
	task, err := MakeTask(&v2alpha1.TaskSpec{
		Library: "common",
		Task:    "exec",
		With: map[string]apiextensionsv1.JSON{
			"cmd": {Raw: b},
		},
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)

	task, err = MakeTask(&v2alpha1.TaskSpec{
		Library: "common",
		Task:    "run",
		With: map[string]apiextensionsv1.JSON{
			"cmd": {Raw: b},
		},
	})
	assert.Nil(t, task)
	assert.Error(t, err)
}
