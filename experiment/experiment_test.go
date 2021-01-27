package experiment

import (
	"testing"

	"github.com/iter8-tools/handler/utils"
	"github.com/stretchr/testify/assert"
)

func TestDryRun(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	err = exp.DryRun()
	assert.NoError(t, err)
}

func TestBuildErrorUnknownTask(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment2.yaml")).Build()
	assert.Error(t, err)
}

func TestBuildErrorGarbageYAML(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment3.yaml")).Build()
	assert.Error(t, err)
}

func TestLocalRunTask(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	err = exp.LocalRun("start", 0)
	assert.NoError(t, err)
}

func TestUnRunnableCommands(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	err = exp.LocalRun("start", -1)
	assert.Error(t, err)
	err = exp.LocalRun("finish", -1)
	assert.Error(t, err)
}

func TestInvalidAction(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment3.yaml")).Build()
	assert.Error(t, err)
}

func TestInvalidExecTask(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment4.yaml")).Build()
	assert.Error(t, err)
}

func TestInvalidActions(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment5.yaml")).Build()
	assert.Error(t, err)
}

func TestMethodsOnNilExperiment(t *testing.T) {
	var e *Experiment
	assert.Error(t, e.extrapolate())
	_, err := e.GetRecommendedBaseline()
	assert.Error(t, err)
	_, err = e.getVersionDetail("noone")
	assert.Error(t, err)
	err = e.DryRun()
	assert.Error(t, err)
	err = e.LocalRun("start", 0)
	assert.Error(t, err)
	_, err = e.getAction("start")
	assert.Error(t, err)
	err = e.Run("start")
	assert.Error(t, err)
}
