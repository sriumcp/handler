package runscript

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/core"
	"github.com/stretchr/testify/assert"
)

func TestMakeFakeRun(t *testing.T) {
	_, err := Make(&v2alpha2.TaskSpec{
		Task: core.StringPointer("fake/fake"),
	})
	assert.Error(t, err)
}

func TestMakeRun(t *testing.T) {
	task, err := Make(&v2alpha2.TaskSpec{
		Run: core.StringPointer("echo hello"),
	})
	assert.NotEmpty(t, task)
	assert.NoError(t, err)
	assert.Equal(t, "echo hello", task.(*Task).With.interpolatedRun)
}

func TestRun(t *testing.T) {
	exp, err := (&core.Builder{}).FromFile(core.CompletePath("../../testdata/common", "runexperiment.yaml")).Build()
	assert.NoError(t, err)
	actionSpec, err := exp.GetActionSpec("start")
	assert.NoError(t, err)

	task, err := Make(&actionSpec[0])
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	err = task.Run(ctx)
	assert.NoError(t, err)
	log.Info(buf.String())
	assert.True(t, strings.Contains(buf.String(), "/quickstart-exp"))

	task, err = Make(&actionSpec[1])
	assert.NoError(t, err)

	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	err = task.Run(ctx)
	assert.NoError(t, err)
	log.Info(buf.String())
	assert.True(t, strings.Contains(buf.String(), "v2"))

	task, err = Make(&actionSpec[2])
	assert.NoError(t, err)

	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	err = task.Run(ctx)
	assert.NoError(t, err)
	log.Info(buf.String())
	assert.True(t, strings.Contains(buf.String(), "v1"))
}
