package gitops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/tasks"
)

const (
	// HelmexUpdateTaskName is the name of the task this file implements
	HelmexUpdateTaskName string = "helmex-update"

	// LocalDir is the directory where the git repo is cloned
	LocalDir string = "/tmp/gitdir"

	// CandidateID is the experiment annotations that uniquely identifies the candidate version.
	CandidateID string = "iter8.candidate.id"

	// DefaultSecretName is the default name of the secret containing the GitHub access token.
	DefaultSecretName string = "ghtoken"

	// DefaultBranch is the default branch to which this task pushes.
	DefaultBranch string = "main"
)

// HelmexUpdateInputs contain the inputs to the helmex-update task to be executed.
type HelmexUpdateInputs struct {
	// GitRepo is the git repo
	GitRepo string `json:"gitRepo" yaml:"gitRepo"`
	// FilePath is the path to values.yaml file within the repo
	FilePath string `json:"filePath" yaml:"filePath"`
	// Username is the name of the GitHub user
	Username string `json:"username" yaml:"username"`
	// Branch is the name of the branch within this repo; default value is master
	Branch *string `json:"branch,omitempty" yaml:"branch,omitempty"`
	// SecretName is the name of the secret containing the GitHub access token
	SecretName *string `json:"secretName,omitempty" yaml:"secretName,omitempty"`
	// SecretName is the namespace of the secret containing the GitHub access token
	SecretNamespace *string `json:"secretNamespace,omitempty" yaml:"secretNamespace,omitempty"`
}

// HelmexUpdateTask enables updates to the values.yaml file within a Helmex git repo.
type HelmexUpdateTask struct {
	tasks.TaskMeta
	With HelmexUpdateInputs `json:"with" yaml:"with"`
}

// MakeHelmexUpdate constructs a HelmexUpdateTask out of a task spec
func MakeHelmexUpdate(t *v2alpha2.TaskSpec) (tasks.Task, error) {
	if t.Task != LibraryName+"/"+HelmexUpdateTaskName {
		return nil, errors.New("library and task need to be " + LibraryName + " and " + HelmexUpdateTaskName)
	}
	var err error
	var jsonBytes []byte
	var bt tasks.Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to HelmexUpdateTask
	if err == nil {
		hut := &HelmexUpdateTask{}
		err = json.Unmarshal(jsonBytes, &hut)
		bt = hut
	}
	return bt, err
}

// initializeDefaults sets default values for HelmexUpdateTaskInputs
func (t *HelmexUpdateTask) initializeDefaults(exp *tasks.Experiment) {
	if t.With.SecretName == nil {
		t.With.SecretName = tasks.StringPointer(DefaultSecretName)
	}
	if t.With.SecretNamespace == nil {
		t.With.SecretNamespace = tasks.StringPointer(exp.Namespace)
	}
	if t.With.Branch == nil {
		t.With.Branch = tasks.StringPointer(DefaultBranch)
	}
}

func (t *HelmexUpdateTask) getToken() (string, error) {
	s, err := tasks.GetSecret(*t.With.SecretNamespace + "/" + *t.With.SecretName)
	if err != nil {
		log.Error(err)
		return "", err
	}
	log.Trace("Got secret")
	token, err := tasks.GetTokenFromSecret(s)
	if err != nil {
		log.Error(err)
		return "", err
	}
	log.Trace("Got token from secret")
	return token, nil
}

// validateInputs to this task
func (t *HelmexUpdateTask) validateInputs() error {
	if !strings.HasPrefix(t.With.GitRepo, "https://") {
		return errors.New("git repo is missing the https:// prefix")
	}
	return nil
}

// updateGitRepoURL updates the git repo url with username and token information
func (t *HelmexUpdateTask) updateGitRepoURL() error {
	token, err := t.getToken()
	if err != nil {
		return err
	}
	t.With.GitRepo = strings.ReplaceAll(t.With.GitRepo, "https://", "https://"+t.With.Username+":"+token+"@")
	return nil
}

// cloneGitRepo locally
// this method is intended to be invoked after initialDefaults()
func (t *HelmexUpdateTask) cloneGitRepo() error {
	// the main idea is to run git clone as a shell command with proper args

	// git clone <repo> <localdir>

	// delete localdir
	cmd := exec.Command("rm", "-rf", LocalDir)
	err := cmd.Run()
	if err != nil {
		log.Error("unable to remove ", LocalDir)
		log.Error(err)
		return err
	}
	// create localdir
	cmd = exec.Command("mkdir", "-p", LocalDir)
	err = cmd.Run()
	if err != nil {
		log.Error("unable to make ", LocalDir)
		log.Error(err)
		return err
	}
	// clone into localdir
	cmd = exec.Command("git", "clone", t.With.GitRepo, LocalDir, "--branch="+*t.With.Branch)
	err = cmd.Run()
	if err != nil {
		log.Error("unable to git clone into ", LocalDir)
		log.Error(err)
	}
	return err
}

// verifyCandidateID ensures that the current experiment is the same one that was created using the values file in the git repo.
// If the candidate id of the values file was updated since the creation of this experiment, this function will return an error.
// It is strongly recommended to change the candidate id in the values file, each time any other field in the candidate section is modified.
// This method is intended to be invoked after cloneGitRepo
func (t *HelmexUpdateTask) verifyCandidateID(exp *tasks.Experiment) error {
	// ensure that the values file can be read
	valuesFilePath := filepath.Join(LocalDir, t.With.FilePath)
	valBytes, err := ioutil.ReadFile(valuesFilePath)
	if err != nil {
		return err
	}
	log.Trace("valBytes: ", string(valBytes))

	// types used for reading in values.yaml into a values struct
	type dynamic struct {
		ID string `json:"id" yaml:"id"`
	}
	type candidate struct {
		Dynamic dynamic `json:"dynamic" yaml:"dynamic"`
	}
	type values struct {
		Candidate candidate `json:"candidate" yaml:"candidate"`
	}

	// pick up and compare candidate id from the values file -- skip if doesn't exist
	vals := values{}
	yaml.Unmarshal(valBytes, &vals)
	log.Trace("Values: ", vals)
	if vals.Candidate.Dynamic.ID != "" {
		if exp.Annotations[CandidateID] != vals.Candidate.Dynamic.ID {
			log.Trace("expID: ", exp.Annotations[CandidateID])
			log.Trace("candidate.dynamic.id: ", vals.Candidate.Dynamic.ID)
			return errors.New("candidate id in experiment does not match the one in values")
		}
	}

	return nil
}

// promoteCandidate updates the values file so that candidate is promoted
func (t *HelmexUpdateTask) promoteCandidate(valuesFilePath string) error {
	// promote candidate
	script := "yq eval '.baseline.dynamic = .candidate.dynamic' -i " + valuesFilePath
	cmd := exec.Command("bash", "-c", script)
	out, err := cmd.CombinedOutput()
	log.Trace("running replacement cmd: ", cmd.String())
	log.Trace("combined output from replacement: ", string(out))
	if err != nil {
		return err
	}

	// nullify candidate
	script = "yq eval '.candidate = null' -i " + valuesFilePath
	cmd = exec.Command("bash", "-c", script)
	out, err = cmd.CombinedOutput()
	log.Trace("running nullify cmd: ", cmd.String())
	log.Trace("combined output from nullify: ", string(out))
	if err != nil {
		return err
	}

	return nil
}

// promoteBaseline updates the values file so that baseline is promoted
func (t *HelmexUpdateTask) promoteBaseline(valuesFilePath string) error {
	// nullify candidate
	script := "yq eval '.candidate = null' -i " + valuesFilePath
	cmd := exec.Command("bash", "-c", script)
	out, err := cmd.CombinedOutput()
	log.Trace("running nullify cmd: ", cmd.String())
	log.Trace("combined output from nullify: ", string(out))
	if err != nil {
		return err
	}

	return nil
}

// updateValuesFile updates the locally cloned values.yaml file.
// this method is intended to be invoked after verifyCandidateID(...)
// ToDo: implement weight updates.
func (t *HelmexUpdateTask) updateValuesFile(exp *tasks.Experiment) error {
	// ensure that the values file can be read
	valuesFilePath := filepath.Join(LocalDir, t.With.FilePath)

	switch exp.Spec.Strategy.TestingPattern {
	case v2alpha2.TestingPatternConformance:
		if exp.Status.Analysis != nil {
			if exp.Status.Analysis.WinnerAssessment != nil {
				if exp.Status.Analysis.WinnerAssessment.Data.WinnerFound {
					return t.promoteCandidate(valuesFilePath)
				}
			}
		}
		return t.promoteBaseline(valuesFilePath)
	case v2alpha2.TestingPatternCanary:
		if exp.Status.VersionRecommendedForPromotion != nil {
			if *exp.Status.VersionRecommendedForPromotion == exp.Spec.VersionInfo.Candidates[0].Name {
				return t.promoteCandidate(valuesFilePath)
			}
		}
		return t.promoteBaseline(valuesFilePath)
	case v2alpha2.TestingPatternAB:
		if exp.Status.VersionRecommendedForPromotion != nil {
			if *exp.Status.VersionRecommendedForPromotion == exp.Spec.VersionInfo.Candidates[0].Name {
				return t.promoteCandidate(valuesFilePath)
			}
		}
		return t.promoteBaseline(valuesFilePath)
	default:
		return errors.New(LibraryName + "/" + HelmexUpdateTaskName + " is currently unsupported with " + string(exp.Spec.Strategy.TestingPattern) + " testing pattern")
	}
}

// updateInGit updates values.yaml file using git push
// this method is intended to be invoked after updateValuesFile(...)
// ToDo: implement request-pr
func (t *HelmexUpdateTask) updateInGit() error {
	script := fmt.Sprintf("cd " + LocalDir + ";" +
		" git config user.email 'iter8@iter8.tools';" +
		" git config user.name '" + t.With.Username + "';" +
		" git commit -a -m 'update values file' --allow-empty;" +
		" git push -f origin " + *t.With.Branch + ";")

	cmd := exec.Command("/bin/bash", "-c", script)
	out, err := cmd.CombinedOutput()
	log.Trace("running script for updating git: ", cmd.String())
	log.Trace("combined output from script: ", string(out))

	return err
}

// Run executes the gitops/helmex-update task
func (t *HelmexUpdateTask) Run(ctx context.Context) error {
	log.Trace("collect task run started...")
	// get experiment from context
	exp, err := tasks.GetExperimentFromContext(ctx)
	if err != nil {
		return err
	}
	t.initializeDefaults(exp)
	err = t.validateInputs()
	if err != nil {
		log.Error("inputs not validated")
		log.Error(err)
		return err
	}
	log.Trace("validated inputts")
	err = t.updateGitRepoURL()
	if err != nil {
		log.Error("unable to update git repo URL")
		log.Error(err)
		return err
	}
	log.Trace("updated git repo url")
	err = t.cloneGitRepo()
	if err != nil {
		log.Error("unable to clone git repo")
		log.Error(err)
		return err
	}
	log.Trace("cloned git repo")
	err = t.verifyCandidateID(exp)
	if err != nil {
		log.Error("unable to verify candidate id")
		log.Error(err)
		return err
	}
	log.Trace("verified candidate id")
	err = t.updateValuesFile(exp)
	if err != nil {
		log.Error("unable to update values file")
		log.Error(err)
		return err
	}
	log.Trace("updated values file")
	err = t.updateInGit()
	if err != nil {
		log.Error("unable to update in Git")
		log.Error(err)
		return err
	}
	log.Trace("updated Git repo")
	return err
}
