package experiment

import (
	"testing"

	"github.com/iter8-tools/handler/utils"
	"github.com/stretchr/testify/assert"
)

func TestDryRun1(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	exp.DryRun()
}

func TestDryRun2(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment2.yaml")).Build()
	assert.Error(t, err)
}

func TestDryRun3(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment2.yaml")).Build()
	assert.Error(t, err)
}

func TestLocalRun1(t *testing.T) {
	_, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment3.yaml")).Build()
	assert.Error(t, err)
}

func TestLocalRun2(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	err = exp.LocalRun("start", 0)
	assert.NoError(t, err)
}

func TestLocalRun3(t *testing.T) {
	exp, err := (&Builder{}).FromFile(utils.CompletePath("../", "testdata/experiment1.yaml")).Build()
	assert.NoError(t, err)
	err = exp.LocalRun("start", -1)
	assert.Error(t, err)
	err = exp.LocalRun("finish", -1)
	assert.Error(t, err)
}
