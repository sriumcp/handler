package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/iter8-tools/etc3/api/v2alpha2"
	"github.com/iter8-tools/handler/core"
	"github.com/iter8-tools/handler/tasks/bash"
	"github.com/iter8-tools/handler/tasks/collect"
	"github.com/iter8-tools/handler/tasks/exec"
	"github.com/iter8-tools/handler/tasks/ghaction"
	"github.com/iter8-tools/handler/tasks/http"
	"github.com/iter8-tools/handler/tasks/readiness"
	"github.com/iter8-tools/handler/tasks/runscript"
	"github.com/iter8-tools/handler/tasks/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/types"
)

// getExperimentNN gets the name and namespace of the experiment from environment variables.
// Returns error if unsuccessful.
func getExperimentNN() (*types.NamespacedName, error) {
	name := viper.GetViper().GetString("experiment_name")
	namespace := viper.GetViper().GetString("experiment_namespace")
	if len(name) == 0 || len(namespace) == 0 {
		return nil, errors.New("invalid experiment name/namespace")
	}
	return &types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, nil
}

// GetAction converts an action spec into an action.
func GetAction(exp *core.Experiment, actionSpec v2alpha2.Action) (core.Action, error) {
	action := make(core.Action, len(actionSpec))
	var err error
	for i := 0; i < len(actionSpec); i++ {
		// if this is a run ... populate runspec
		if core.IsARun(&actionSpec[i]) {
			if action[i], err = runscript.Make(&actionSpec[i]); err != nil {
				break
			}
		} else if core.IsATask(&actionSpec[i]) {
			if action[i], err = MakeTask(&actionSpec[i]); err != nil {
				break
			}
		} else {
			return nil, errors.New("action spec contains item that is neither run spec nor task spec")
		}
	}
	return action, err
}

// run is a helper function used in the definition of runCmd cobra command.
func run(cmd *cobra.Command, args []string) error {
	nn, err := getExperimentNN()
	if err == nil {
		var exp *core.Experiment
		// if localExperiment is an empty string
		// get experiment from cluster ...
		// else if localExperiment is a non-empty string (should be a valid path to an experimnent file)
		// get experiment from local file
		if exp, err = (&core.Builder{}).FromCluster(nn).Build(); err == nil {
			var actionSpec v2alpha2.Action
			if actionSpec, err = exp.GetActionSpec(action); err == nil {
				var action core.Action
				if action, err = GetAction(exp, actionSpec); err == nil {
					ctx := context.WithValue(context.Background(), core.ContextKey("experiment"), exp)
					log.Trace("created context for experiment")
					err = action.Run(ctx)
					if err == nil {
						return nil
					}
				}
			} else {
				log.Error("could not find specified action: " + action)
				return nil
			}
		}
	}
	return err
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run an action",
	Long:  `Sequentially execute all tasks in the specified action; if any task run results in an error, exit immediately with error. Run can optionally use a local experiment manifest. Only a subset of tasks (including metrics collection) are supported for local experiments.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := run(cmd, args); err != nil {
			log.Error("Exiting with error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringVarP(&action, "action", "a", "", "name of the action")
	runCmd.PersistentFlags().StringVarP(&localExperiment, "experiment", "e", "", "full path to the local experiment manifest; optional")
	runCmd.MarkPersistentFlagRequired("action")
}

// MakeTask constructs a Task from a TaskSpec or returns an error if any.
func MakeTask(t *v2alpha2.TaskSpec) (core.Task, error) {
	if t == nil || t.Task == nil || len(*t.Task) == 0 {
		return nil, errors.New("nil or empty task found")
	}
	switch *t.Task {
	case bash.TaskName:
		return bash.Make(t)
	case collect.TaskName:
		return collect.Make(t)
	case exec.TaskName:
		return exec.Make(t)
	case ghaction.TaskName:
		return ghaction.Make(t)
	case http.TaskName:
		return http.Make(t)
	case readiness.TaskName:
		return readiness.Make(t)
	case slack.TaskName:
		return slack.Make(t)
	default:
		return nil, errors.New("unknown task: " + *t.Task)
	}
}
