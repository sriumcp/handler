package runscript

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/core"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = core.GetLogger()
}

// Inputs for the run task may contain a secret reference
type Inputs struct {
	Secret          *string `json:"secret" yaml:"secret"`
	interpolatedRun string
}

// Task encapsulates a command that can be executed.
type Task struct {
	core.TaskMeta `json:",inline" yaml:",inline"`
	With          Inputs `json:"with" yaml:"with"`
}

// Make converts an run spec into a run.
func Make(t *v2alpha2.TaskSpec) (core.Task, error) {
	if !core.IsARun(t) {
		return nil, fmt.Errorf("invalid run spec")
	}
	var jsonBytes []byte
	var task Task
	// convert t to jsonBytes
	jsonBytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	// convert jsonString to ExecTask
	task = Task{}
	err = json.Unmarshal(jsonBytes, &task)
	return &task, err
}

// EnhancedExperiment supports enhanced interpolation behaviors
type EnhancedExperiment struct {
	*core.Experiment
}

// Interpolate the script.
func (t *Task) Interpolate(ctx context.Context) error {
	exp, err := core.GetExperimentFromContext(ctx)
	if err != nil {
		log.Error(err)
		return err
	}
	ee := EnhancedExperiment{Experiment: exp}
	log.Trace("experiment", exp)

	var templ *template.Template
	if templ, err = template.New("templated script").Parse(*t.TaskMeta.Run); err == nil {
		buf := bytes.Buffer{}
		if err = templ.Execute(&buf, ee); err == nil {
			t.With.interpolatedRun = buf.String()
			return nil
		}
		log.Error("template execution error: ", err)
		return errors.New("cannot interpolate string")
	}
	log.Error("template creation error: ", err)
	return errors.New("cannot interpolate string")
}

// Run the command.
func (t *Task) Run(ctx context.Context) error {
	err := t.Interpolate(ctx)
	if err != nil {
		log.Error(err)
		return err
	}

	cmd := exec.Command("/bin/bash", "-c", t.With.interpolatedRun)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Running task: " + cmd.String())
	return cmd.Run()
}
