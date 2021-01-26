package def

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/iter8-tools/handler/base"
)

// ExecInputs contain the name and arguments of the command to be executed.
type ExecInputs struct {
	Cmd  string        `json:"cmd" yaml:"cmd"`
	Args []interface{} `json:"args,omitempty" yaml:"args,omitempty"`
}

// ExecTask struct enables JSON serialization and deserialization of a command to be executed.
type ExecTask struct {
	Library string     `json:"library" yaml:"library"`
	Task    string     `json:"task" yaml:"task"`
	With    ExecInputs `json:"with" yaml:"with"`
}

// Run the command.
func (t *ExecTask) Run(ctx context.Context) error {
	args := make([]string, len(t.With.Args))
	for i := 0; i < len(args); i++ {
		args[i] = fmt.Sprint(t.With.Args[i])
	}
	cmd := exec.Command(t.With.Cmd, args...)
	log.Info("Running task: " + cmd.String())
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// DryRun explains this command.
func (t *ExecTask) DryRun() {
	args := make([]string, len(t.With.Args))
	for i := 0; i < len(args); i++ {
		args[i] = fmt.Sprint(t.With.Args[i])
	}
	cmd := exec.Command(t.With.Cmd, args...)
	fmt.Println("  - Execute command: ", cmd.String())
}

// Extrapolate each argument.
func (t *ExecTask) Extrapolate(tags *base.Tags) error {
	for i := 0; i < len(t.With.Args); i++ {
		var err error
		str := fmt.Sprint(t.With.Args[i])
		if t.With.Args[i], err = tags.Extrapolate(&str); err != nil {
			return errors.New("unable to extrapolate exec task")
		}
	}
	return nil
}

// MakeExec makes an empty exec task.
func MakeExec(t *base.TaskMeta) (base.Task, error) {
	if t.Library != "default" || t.Task != "exec" {
		return nil, errors.New("library and task need to be 'default' and 'exec'")
	}
	return &ExecTask{}, nil
}
