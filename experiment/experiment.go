// Package experiment enables construction of an experiment object with handler/task lists within it.
package experiment

import (
	"encoding/json"
	"errors"

	iter8 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/base"
	"github.com/iter8-tools/handler/lib/def"
	"github.com/iter8-tools/handler/lib/knative"
	"github.com/iter8-tools/handler/utils"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = utils.GetLogger()
}

// Experiment is an enhancement of v2alpha1.Experiment struct that contains task list information.
type Experiment struct {
	iter8.Experiment
	Spec Spec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

// Spec is an enhancement of v2alpha1.ExperimentSpec struct that contains task list information.
type Spec struct {
	iter8.ExperimentSpec
	Strategy Strategy `json:"strategy" yaml:"strategy"`
}

// Strategy is an enhancement of v2alpha1.Strategy struct that contains task list information.
type Strategy struct {
	iter8.Strategy
	Handlers *Handlers `json:"handlers,omitempty" yaml:"handlers,omitempty"`
}

// Handlers is an enhancement of v2alpha1.Handlers struct that contains task list information.
type Handlers struct {
	iter8.Handlers
	// Map of task lists.
	Actions *ActionMap `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// ActionMap type represents a map whose keys are actions names, and whose values are actions.
type ActionMap map[string]*Action

// Action is a slice of Tasks.
type Action []base.Task

// UnmarshalJSON builds an ActionMap from a byte slice.
func (sm *ActionMap) UnmarshalJSON(data []byte) error {
	actionMapRaw := make(map[string][]json.RawMessage)
	var err error
	if err = json.Unmarshal(data, &actionMapRaw); err == nil {
		actions := make(ActionMap)
		// first create raw actions, then extract TaskMeta from then, and then extract tasks
		for actionName, rawTasksForAction := range actionMapRaw {
			action := make(Action, len(rawTasksForAction))
			actions[actionName] = &action
			for i := 0; i < len(rawTasksForAction); i++ {
				taskMeta := &base.TaskMeta{}
				if err = json.Unmarshal(rawTasksForAction[i], taskMeta); err == nil {
					// demux library here...
					var makeTask func(t *base.TaskMeta) (base.Task, error)
					switch taskMeta.Library {
					case "default":
						makeTask = def.MakeTask
					case "knative":
						makeTask = knative.MakeTask
					default:
						return errors.New("Unrecognized task library")
					}
					var task base.Task
					if task, err = makeTask(taskMeta); err == nil {
						if err = json.Unmarshal(rawTasksForAction[i], task); err == nil {
							action[i] = task
							log.Trace("using... ")
						} else {
							taskBytes, _ := json.MarshalIndent(task, "", "  ")
							log.Error("cannot unmarshal task: " + string(taskBytes))
							log.Error(err)
							return err
						}
					} else {
						log.Error("cannot make task: ", *taskMeta)
						return err
					}
				}
			}
		}
		if err != nil {
			log.Error("cannot unmarshal ActionMap")
			return errors.New("cannot unmarshal ActionMap")
		}
		*sm = actions
		return nil
	}
	return errors.New("cannot unmarshal ActionMap")
}

// Builder helps in construction of an experiment.
type Builder struct {
	err error
	exp *Experiment
}

// Build returns the built experiment or error.
// Must call FromFile or FromCluster on b prior to invoking Build.
func (b *Builder) Build() (*Experiment, error) {
	log.Trace(b)
	return b.exp, b.err
}
