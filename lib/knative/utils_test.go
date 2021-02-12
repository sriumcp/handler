package knative

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"github.com/iter8-tools/handler/experiment"
	"github.com/iter8-tools/handler/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetNamespacedNameForKsvc(t *testing.T) {
	var err error
	var exp *experiment.Experiment
	var nn *types.NamespacedName
	exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../../", "testdata/experiment6.yaml")).Build()
	assert.NoError(t, err)

	nn, err = GetNamespacedNameForKsvc(exp)
	assert.Equal(t, *nn, types.NamespacedName{
		Namespace: "default",
		Name:      "sklearn-iris",
	})
	assert.NoError(t, err)

	exp, err = (&experiment.Builder{}).FromFile(utils.CompletePath("../../", "testdata/experiment2.yaml")).Build()
	assert.NoError(t, err)

	nn, err = GetNamespacedNameForKsvc(exp)
	assert.Nil(t, nn)
	assert.Error(t, err)
}
