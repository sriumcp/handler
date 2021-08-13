package runscript

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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
	Secret *string `json:"secret" yaml:"secret"`
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

// Run the command.
func (t *Task) Run(ctx context.Context) error {
	cmd := exec.Command("/bin/bash", "-c", *t.TaskMeta.Run)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Info("Running task: " + cmd.String())
	log.Trace(*t.TaskMeta.Run)
	return cmd.Run()
}
