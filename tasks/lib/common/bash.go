package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/tasks"
)

const (
	// BashTaskName is the name of the bash task
	BashTaskName string = "bash"
)

// BashInputs contain the name and arguments of the command to be executed.
type BashInputs struct {
	VersionInfo []tasks.VersionInfo `json:"versionInfo,omitempty" yaml:"versionInfo,omitempty"`
	Script      string              `json:"script" yaml:"script"`
}

// BashTask encapsulates a command that can be executed.
type BashTask struct {
	tasks.TaskMeta `json:",inline" yaml:",inline"`
	With           BashInputs `json:"with" yaml:"with"`
}

// MakeBashTask converts an exec task spec into an exec task.
func MakeBashTask(t *v2alpha2.TaskSpec) (tasks.Task, error) {
	if t.Task != LibraryName+"/"+BashTaskName {
		return nil, fmt.Errorf("library and task need to be '%s' and '%s'", LibraryName, BashTaskName)
	}
	var jsonBytes []byte
	var task tasks.Task
	// convert t to jsonBytes
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	// convert jsonString to ExecTask
	task = &BashTask{}
	err = json.Unmarshal(jsonBytes, &task)
	return task, err
}

// Run the command.
func (t *BashTask) Run(ctx context.Context) error {
	exp, err := tasks.GetExperimentFromContext(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Trace("experiment", exp)

	obj, err := exp.ToMap()
	if err != nil {
		// error already logged by ToMap()
		// don't log it again
		return err
	}

	// prepare for interpolation; add experiment as tag
	// Note that if versionRecommendedForPromotion is not set or there is no version corresponding to it,
	// then some placeholders may not be replaced
	tags := tasks.NewTags().
		With("this", obj).
		WithRecommendedVersionForPromotion(&exp.Experiment, t.With.VersionInfo)

	// interpolate - replaces placeholders in the script with values
	script, err := tags.Interpolate(&t.With.Script)

	log.Trace(script)
	args := []string{"-c", script}

	cmd := exec.Command("/bin/bash", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Running task: " + cmd.String())
	log.Trace(args)
	err = cmd.Run()

	return err
}
