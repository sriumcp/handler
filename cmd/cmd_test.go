package cmd

import (
	"os"
	"testing"

	"github.com/iter8-tools/handler/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	log = utils.GetLogger()
}

func TestVersion(t *testing.T) {
	versionCmd.Run(nil, nil)
}

func TestInitConfig(t *testing.T) {
	initConfig()
}

func TestInitConfigEmptyCfgFile(t *testing.T) {
	cfgFile = ""
	initConfig()
}

func TestEnv(t *testing.T) {
	os.Setenv("EXPERIMENT_NAME", "name")
	os.Setenv("EXPERIMENT_NAMESPACE", "namespace")
	nn, err := getExperimentNN()
	assert.Equal(t, "name", nn.Name)
	assert.Equal(t, "namespace", nn.Namespace)
	assert.NoError(t, err)

	os.Unsetenv("EXPERIMENT_NAME")
	os.Unsetenv("EXPERIMENT_NAMESPACE")
	_, err = getExperimentNN()
	assert.Error(t, err)

	os.Setenv("EXPERIMENT_NAMESPACE", "namespace")
	nn, err = getExperimentNN()
	assert.Equal(t, "namespace", nn.Namespace)
	assert.Error(t, err)
}
