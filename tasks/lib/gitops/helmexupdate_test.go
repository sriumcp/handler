package gitops

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/tasks"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestMakeTask(t *testing.T) {
	hut := HelmexUpdateTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    HelmexUpdateTaskName,
		},
		With: HelmexUpdateInputs{
			GitRepo:  "https://my.repo",
			FilePath: "myfile.yaml",
		},
	}

	repoJSON, _ := json.Marshal("https://my.repo")
	fileJSON, _ := json.Marshal("myfile.yaml")

	// good inputs
	ts := v2alpha2.TaskSpec{
		Task: LibraryName + "/" + HelmexUpdateTaskName,
		With: map[string]v1.JSON{
			"gitRepo":  {Raw: repoJSON},
			"filePath": {Raw: fileJSON},
		},
	}
	hut1, err := MakeTask(&ts)
	assert.Equal(t, hut.With, hut1.(*HelmexUpdateTask).With)
	assert.NoError(t, err)

	// erroneous library name
	ts = v2alpha2.TaskSpec{
		Task: "non-existent-library/" + HelmexUpdateTaskName,
		With: map[string]v1.JSON{
			"gitRepo":  {Raw: repoJSON},
			"filePath": {Raw: fileJSON},
		},
	}
	_, err = MakeTask(&ts)
	assert.Error(t, err)
}

func TestInitializeDefaults(t *testing.T) {
	hut := HelmexUpdateTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    HelmexUpdateTaskName,
		},
		With: HelmexUpdateInputs{
			GitRepo:  "",
			FilePath: "",
		},
	}
	exp, err := (&tasks.Builder{}).FromFile(tasks.CompletePath("../../../", "testdata/helmex-update/experiment.yaml")).Build()
	assert.NoError(t, err)

	hut.initializeDefaults(exp)
	assert.Equal(t, DefaultSecretName, *hut.With.SecretName)
	assert.Equal(t, exp.Namespace, *hut.With.SecretNamespace)
	assert.Equal(t, DefaultBranch, *hut.With.Branch)
}

func TestValidateInputs(t *testing.T) {
	hut := HelmexUpdateTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    HelmexUpdateTaskName,
		},
		With: HelmexUpdateInputs{
			GitRepo:  "https://github.com/iter8-tools/iter8.git",
			FilePath: "values.yaml",
			Branch:   tasks.StringPointer("master"),
		},
	}
	err := hut.validateInputs()
	assert.NoError(t, err)

	hut = HelmexUpdateTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    HelmexUpdateTaskName,
		},
		With: HelmexUpdateInputs{
			GitRepo:  "git@github.com/iter8-tools/iter8.git",
			FilePath: "values.yaml",
			Branch:   tasks.StringPointer("master"),
		},
	}
	err = hut.validateInputs()
	assert.Error(t, err)
}

func TestCloneGitRepo(t *testing.T) {
	hut := HelmexUpdateTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    HelmexUpdateTaskName,
		},
		With: HelmexUpdateInputs{
			GitRepo:  "https://github.com/iter8-tools/iter8.git",
			FilePath: "",
			Branch:   tasks.StringPointer("master"),
		},
	}
	exp, err := (&tasks.Builder{}).FromFile(tasks.CompletePath("../../../", "testdata/helmex-update/experiment.yaml")).Build()
	assert.NoError(t, err)

	hut.initializeDefaults(exp)
	err = hut.cloneGitRepo()
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(LocalDir, "LICENSE"))
	assert.NoError(t, err)
}

func TestVerifyCandidateID(t *testing.T) {
	hut := HelmexUpdateTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    HelmexUpdateTaskName,
		},
		With: HelmexUpdateInputs{
			GitRepo:  "https://github.com/iter8-tools/iter8.git",
			FilePath: "samples/second-exp/values.yaml",
			Branch:   tasks.StringPointer("master"),
		},
	}

	exp, err := (&tasks.Builder{}).FromFile(tasks.CompletePath("../../../", "testdata/helmex-update/experiment.yaml")).Build()
	assert.NoError(t, err)

	hut.initializeDefaults(exp)
	err = hut.cloneGitRepo()
	assert.NoError(t, err)

	// verify candidate id
	err = hut.verifyCandidateID(exp)
	assert.NoError(t, err)

	// mess up the id in the experiment ... should fail
	original := exp.Annotations[CandidateID]
	exp.Annotations[CandidateID] = "no-way-this-matches-the-correct-id"
	err = hut.verifyCandidateID(exp)
	assert.Error(t, err)
	exp.Annotations[CandidateID] = original

	// mess up the id in the values file ... should fail
	localValFile := filepath.Join(LocalDir, hut.With.FilePath)

	// do a file replace
	input, err := ioutil.ReadFile(localValFile)
	assert.NoError(t, err)
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, original) {
			lines[i] = strings.ReplaceAll(lines[i], original, "no-way-this-works")
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(localValFile, []byte(output), 0664)
	assert.NoError(t, err)

	// no make sure candidate verification fails
	err = hut.verifyCandidateID(exp)
	assert.Error(t, err)
}

// types used for reading in values.yaml into a values struct
type dynamic struct {
	ID string `json:"id" yaml:"id"`
}
type version struct {
	Dynamic dynamic `json:"dynamic" yaml:"dynamic"`
}
type values struct {
	Baseline  *version `json:"baseline" yaml:"baseline"`
	Candidate *version `json:"candidate" yaml:"candidate"`
}

func tcpro(t *testing.T, localValFile string, exp *tasks.Experiment, hutPtr *HelmexUpdateTask) {
	// read in the old values file and store it ...
	oldVal := values{}
	oldValBytes, err := ioutil.ReadFile(localValFile)
	assert.NoError(t, err)
	err = yaml.Unmarshal(oldValBytes, &oldVal)
	assert.NoError(t, err)

	// update values file
	err = hutPtr.updateValuesFile(exp)
	assert.NoError(t, err)

	// read in the new values file and store it ...
	newVal := values{}
	newValBytes, err := ioutil.ReadFile(localValFile)
	assert.NoError(t, err)
	err = yaml.Unmarshal(newValBytes, &newVal)
	assert.NoError(t, err)

	// do more assertions now ...
	assert.Equal(t, oldVal.Candidate.Dynamic, newVal.Baseline.Dynamic)
	assert.Empty(t, newVal.Candidate)
}

func tbpro(t *testing.T, localValFile string, exp *tasks.Experiment, hutPtr *HelmexUpdateTask) {
	// read in the old values file and store it ...
	oldVal := values{}
	oldValBytes, err := ioutil.ReadFile(localValFile)
	assert.NoError(t, err)
	err = yaml.Unmarshal(oldValBytes, &oldVal)
	assert.NoError(t, err)

	// update values file
	err = hutPtr.updateValuesFile(exp)
	assert.NoError(t, err)

	// read in the new values file and store it ...
	newVal := values{}
	newValBytes, err := ioutil.ReadFile(localValFile)
	assert.NoError(t, err)
	err = yaml.Unmarshal(newValBytes, &newVal)
	assert.NoError(t, err)

	// do more assertions now ...
	assert.Equal(t, oldVal.Baseline.Dynamic, newVal.Baseline.Dynamic)
	assert.Empty(t, newVal.Candidate)
}

func duplicateValuesFile(original string, duplicate string) {
	cmd := exec.Command("cp", original, duplicate)
	cmd.Run()
}

func restoreValuesFile(original string, current string) {
	cmd := exec.Command("cp", original, current)
	cmd.Run()
}
func TestUpdateValuesFile(t *testing.T) {
	hut := HelmexUpdateTask{
		TaskMeta: tasks.TaskMeta{
			Library: LibraryName,
			Task:    HelmexUpdateTaskName,
		},
		With: HelmexUpdateInputs{
			GitRepo:  "https://github.com/iter8-tools/iter8.git",
			FilePath: "samples/second-exp/values.yaml",
			Branch:   tasks.StringPointer("master"),
		},
	}

	exp, err := (&tasks.Builder{}).FromFile(tasks.CompletePath("../../../", "testdata/helmex-update/experiment.yaml")).Build()
	assert.NoError(t, err)

	hut.initializeDefaults(exp)
	err = hut.cloneGitRepo()
	assert.NoError(t, err)

	// verify candidate id
	err = hut.verifyCandidateID(exp)
	assert.NoError(t, err)

	// local values file
	localValFile := filepath.Join(LocalDir, hut.With.FilePath)
	originalValFile := localValFile + ".orig"

	/*
		Conformance with baseline promotion
	*/
	duplicateValuesFile(localValFile, originalValFile)
	tbpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	/*
		Conformance with candidate promotion
	*/
	duplicateValuesFile(localValFile, originalValFile)
	exp.Status.Analysis = &v2alpha2.Analysis{}
	exp.Status.Analysis.WinnerAssessment = &v2alpha2.WinnerAssessmentAnalysis{}
	exp.Status.Analysis.WinnerAssessment.Data.WinnerFound = true

	duplicateValuesFile(localValFile, originalValFile)
	tcpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	// Canary and A/B experiments
	exp.Spec.VersionInfo.Candidates = []v2alpha2.VersionDetail{{
		Name: "latest-version",
	}}

	// Canary
	exp.Spec.Strategy.TestingPattern = v2alpha2.TestingPatternCanary

	/*
		Canary with baseline promotion
	*/
	duplicateValuesFile(localValFile, originalValFile)
	exp.Status.VersionRecommendedForPromotion = nil
	tbpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	duplicateValuesFile(localValFile, originalValFile)
	exp.Status.VersionRecommendedForPromotion = tasks.StringPointer(exp.Spec.VersionInfo.Baseline.Name)
	tbpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	/*
		Canary with candidate promotion
	*/
	duplicateValuesFile(localValFile, originalValFile)
	exp.Status.VersionRecommendedForPromotion = tasks.StringPointer(exp.Spec.VersionInfo.Candidates[0].Name)
	tcpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	// A/B
	exp.Spec.Strategy.TestingPattern = v2alpha2.TestingPatternAB

	/*
		A/B with baseline promotion
	*/
	duplicateValuesFile(localValFile, originalValFile)
	exp.Status.VersionRecommendedForPromotion = nil
	tbpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	duplicateValuesFile(localValFile, originalValFile)
	exp.Status.VersionRecommendedForPromotion = tasks.StringPointer(exp.Spec.VersionInfo.Baseline.Name)
	tbpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	/*
		A/B with candidate promotion
	*/
	duplicateValuesFile(localValFile, originalValFile)
	exp.Status.VersionRecommendedForPromotion = tasks.StringPointer(exp.Spec.VersionInfo.Candidates[0].Name)
	tcpro(t, localValFile, exp, &hut)
	restoreValuesFile(originalValFile, localValFile)

	// A/B/n
	exp.Spec.Strategy.TestingPattern = v2alpha2.TestingPatternABN
	err = hut.updateValuesFile(exp)
	assert.Error(t, err)

}
