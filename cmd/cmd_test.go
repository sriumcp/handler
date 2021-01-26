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
func TestDryRun1(t *testing.T) {
	filePath = utils.CompletePath("../", "testdata/experiment1.yaml")
	dryrunCmd.Run(nil, nil)
}

func TestDryRun2(t *testing.T) {
	filePath = utils.CompletePath("../", "testdata/experiment3.yaml")
	dryrunCmd.Run(nil, nil)
}

func TestLocalRun3(t *testing.T) {
	filePath = utils.CompletePath("../", "testdata/experiment3.yaml")
	localrunCmd.Run(nil, nil)
}

func TestLocalRun4(t *testing.T) {
	filePath = utils.CompletePath("../", "testdata/experiment2.yaml")
	localrunCmd.Run(nil, nil)
}

func TestLocalRun5(t *testing.T) {
	action = "start"
	task = 10
	filePath = utils.CompletePath("../", "testdata/experiment1.yaml")
	localrunCmd.Run(nil, nil)
	task = 1
	localrunCmd.Run(nil, nil)
}

func TestVersion(t *testing.T) {
	versionCmd.Run(nil, nil)
}

func TestExecute(t *testing.T) {
	Execute()
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
	name, namespace, err := getExperimentNN()
	assert.Equal(t, "name", name)
	assert.Equal(t, "namespace", namespace)
	assert.NoError(t, err)

	os.Unsetenv("EXPERIMENT_NAME")
	os.Unsetenv("EXPERIMENT_NAMESPACE")
	name, namespace, err = getExperimentNN()
	assert.Error(t, err)

	os.Setenv("EXPERIMENT_NAMESPACE", "namespace")
	name, namespace, err = getExperimentNN()
	assert.Equal(t, "namespace", namespace)
	assert.Error(t, err)

}
