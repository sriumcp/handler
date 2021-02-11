package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/iter8-tools/handler/base"
	"github.com/iter8-tools/handler/experiment"
)

// ExecInputs contain the name and arguments of the command to be executed.
type ExecInputs struct {
	Cmd  string        `json:"cmd" yaml:"cmd"`
	Args []interface{} `json:"args,omitempty" yaml:"args,omitempty"`
}

// ExecTask encapsulates a command that can be executed.
type ExecTask struct {
	Library string     `json:"library" yaml:"library"`
	Task    string     `json:"task" yaml:"task"`
	With    ExecInputs `json:"with" yaml:"with"`
}

// Run the command.
func (t *ExecTask) Run(ctx context.Context) error {
	exp, err := experiment.GetExperimentFromContext(ctx)
	if err == nil {
		inputArgs := make([]string, len(t.With.Args))
		for i := 0; i < len(inputArgs); i++ {
			inputArgs[i] = fmt.Sprint(t.With.Args[i])
		}
		var args []string
		if args, err = exp.Extrapolate(inputArgs); err == nil {
			cmd := exec.Command(t.With.Cmd, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Info("Running task: " + cmd.String())
			err = cmd.Run()
		}
	}
	return err
}

// MakeExec converts an exec task spec into an exec task.
func MakeExec(t *base.TaskSpec) (base.Task, error) {
	if t.Library != "common" || t.Task != "exec" {
		return nil, errors.New("library and task need to be 'common' and 'exec'")
	}
	var err error
	var jsonBytes []byte
	var et base.Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to ExecTask
	if err == nil {
		et = &ExecTask{}
		err = json.Unmarshal(jsonBytes, &et)
	}
	return et, err
}
