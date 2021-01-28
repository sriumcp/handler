package experiment

import (
	"encoding/json"
	"testing"

	iter8 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/utils"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
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
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment4.yaml")).Build()
	jm, _ := json.MarshalIndent(exp, "", "  ")
	log.Trace("experiment with invalid exec task:", string(jm))
	assert.Error(t, err)
}

func TestInvalidActions(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment5.yaml")).Build()
	assert.Error(t, err)
}

func TestInvalidTaskMeta(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment8.yaml")).Build()
	assert.Error(t, err)
}

func TestStringAction(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment9.yaml")).Build()
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
	_, err = e.GetAction("start")
	assert.Error(t, err)
	err = e.Run("start")
	assert.Error(t, err)
}

func TestExtrapolateWithoutHandlerStanza(t *testing.T) {
	var e *Experiment = &Experiment{}
	assert.NoError(t, e.extrapolate())

	_, err := e.GetAction("start")
	assert.Error(t, err)

	e.Spec.Strategy.Handlers = &Handlers{}
	_, err = e.GetAction("start")
	assert.Error(t, err)
}

func TestDryRunWithoutHandlerStanza(t *testing.T) {
	var e *Experiment = &Experiment{}
	assert.NoError(t, e.DryRun())

	e.Spec.Strategy.Handlers = &Handlers{}
	assert.NoError(t, e.DryRun())
}

func TestDryRunWithBadExtrapolation(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)

	badBaseline := ""
	exp.Status.RecommendedBaseline = &badBaseline
	assert.Error(t, exp.extrapolate())
	assert.NoError(t, exp.DryRun())

	exp.Status.RecommendedBaseline = nil
	assert.Error(t, exp.extrapolate())
	assert.NoError(t, exp.DryRun())
}

func TestExperimentWithNilAction(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment10.yaml")).Build()
	assert.NoError(t, err)

	assert.NoError(t, exp.DryRun())
}

func TestExtrapolateWithoutRecommendedBaseline(t *testing.T) {
	var e *Experiment = &Experiment{
		Experiment: *iter8.NewExperiment("default", "default").Build(),
	}
	e.Spec.Strategy.Handlers = &Handlers{
		Actions: &ActionMap{},
	}
	assert.Error(t, e.extrapolate())
}

func TestExtrapolateWithoutVersionTags(t *testing.T) {
	var iter8Experiment = *iter8.NewExperiment("some", "exp").Build()
	spec := Spec{}
	spec.VersionInfo = &iter8.VersionInfo{
		Baseline: iter8.VersionDetail{
			Name:         "default",
			Tags:         nil,
			WeightObjRef: &v1.ObjectReference{},
		},
		Candidates: []iter8.VersionDetail{
			{
				Name:         "canary",
				Tags:         nil,
				WeightObjRef: &v1.ObjectReference{},
			},
		},
	}
	var e *Experiment = &Experiment{
		Experiment: iter8Experiment,
		Spec:       spec,
	}
	x := "default"
	e.Status.RecommendedBaseline = &x

	e.Spec.Strategy.Handlers = &Handlers{
		Actions: &ActionMap{},
	}
	assert.NoError(t, e.extrapolate())
}

func TestGetVersionInfo(t *testing.T) {
	var iter8Experiment = *iter8.NewExperiment("some", "exp").Build()
	spec := Spec{}
	spec.VersionInfo = &iter8.VersionInfo{
		Baseline: iter8.VersionDetail{
			Name:         "default",
			Tags:         nil,
			WeightObjRef: &v1.ObjectReference{},
		},
		Candidates: []iter8.VersionDetail{
			{
				Name:         "canary",
				Tags:         nil,
				WeightObjRef: &v1.ObjectReference{},
			},
		},
	}
	var e *Experiment = &Experiment{
		Experiment: iter8Experiment,
		Spec:       spec,
	}
	x := "default"
	e.Status.RecommendedBaseline = &x
	_, err := e.getVersionDetail("default")
	assert.NoError(t, err)

	e.Spec.Strategy.Handlers = &Handlers{
		Actions: &ActionMap{},
	}
	assert.NoError(t, e.extrapolate())

	_, err = e.getVersionDetail("canary")
	assert.NoError(t, err)

	_, err = e.getVersionDetail("random")
	assert.Error(t, err)
}
