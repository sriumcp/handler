// Package experiment enables construction of an experiment object with handler/task lists within it.
package experiment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
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
							log.Trace("using.. ")
						} else {
							taskBytes, _ := json.MarshalIndent(task, "", "  ")
							log.Error("cannot unmarshal task: ", string(taskBytes))
						}
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

// FromFile method builds an experiment from a yaml file.
func (b *Builder) FromFile(filePath string) *Builder {
	var err error
	var data []byte
	if data, err = ioutil.ReadFile(filePath); err == nil {
		exp := &Experiment{}
		if err = yaml.Unmarshal(data, exp); err == nil {
			actions, _ := json.MarshalIndent(exp.Spec.Strategy.Handlers.Actions, "", "  ")
			log.Trace(string(actions))
			if err = exp.extrapolate(); err == nil {
				b.exp = exp
				return b
			}
		} else {
			log.Error(err)
		}
	}
	log.Error("cannot extrapolate experiment")
	b.err = errors.New("cannot extrapolate experiment")
	return b
}

func (e *Experiment) extrapolate() (er error) {
	if e.Spec.Strategy.Handlers == nil || e.Spec.Strategy.Handlers.Actions == nil {
		return nil
	}
	if rb, err := e.GetRecommendedBaseline(); err == nil {
		if version, err := e.getVersionDetail(rb); err == nil {
			if version.Tags == nil {
				return nil
			}
			for _, action := range *e.Spec.Strategy.Handlers.Actions {
				for i := 0; i < len(*action); i++ {
					log.Trace(i, *action, version.Tags)
					err = (*action)[i].Extrapolate(&base.Tags{M: version.Tags})
				}
			}
		}
	} else {
		log.Warn("cannot get recommended baseline")
		return nil
	}
	return nil
}

// GetRecommendedBaseline returns the next baseline recommended in the experiment.
func (e *Experiment) GetRecommendedBaseline() (string, error) {
	if e.Status.RecommendedBaseline == nil {
		return "", errors.New("Recommended baseline not found in experiment status")
	}
	return *e.Status.RecommendedBaseline, nil
}

// get VersionDetail given a version name.
func (e *Experiment) getVersionDetail(versionName string) (*iter8.VersionDetail, error) {
	if e.Spec.VersionInfo != nil {
		if e.Spec.VersionInfo.Baseline.Name == versionName {
			return &e.Spec.VersionInfo.Baseline, nil
		}
		for i := 0; i < len(e.Spec.VersionInfo.Candidates); i++ {
			if e.Spec.VersionInfo.Candidates[i].Name == versionName {
				return &e.Spec.VersionInfo.Candidates[i], nil
			}
		}
	}
	return nil, errors.New("no version found with name " + versionName)
}

// Build returns the built experiment or error.
// Must call FromFile or FromCluster prior to invoking Build.
func (b *Builder) Build() (*Experiment, error) {
	log.Trace(b)
	return b.exp, b.err
}

// DryRun explains the tasks in this experiment.
func (e *Experiment) DryRun() {
	handlers := e.Spec.Strategy.Handlers
	if handlers == nil {
		fmt.Println("Experiment does not have a handler stanza")
		return
	}
	actions := handlers.Actions
	if actions == nil {
		fmt.Println("Experiment does not have a map of actions")
		return
	}
	fmt.Println("-------------------")
	fmt.Printf("Experiment has %d actions. ", len(*actions))
	keys := make([]string, len(*actions))
	i := 0
	for k := range *actions {
		keys[i] = k
		i++
	}
	fmt.Printf("Action names: %s\n", keys)
	for j := 0; j < i; j++ {
		name := keys[j]
		fmt.Printf("- Action %s performs the following tasks\n", name)
		action := (*actions)[name]
		log.Trace("Action ptr: ", action)
		action.DryRun()
	}
}

// DryRun explains the tasks in this action.
func (a *Action) DryRun() {
	for i := 0; i < len(*a); i++ {
		task := (*a)[i]
		log.Trace("Task ptr: ", task)
		task.DryRun()
	}
}

func (e *Experiment) getAction(name string) (*Action, error) {
	if e == nil {
		return nil, errors.New("nil experiment")
	}
	if e.Spec.Strategy.Handlers == nil {
		return nil, errors.New("nil handlers")
	}
	if e.Spec.Strategy.Handlers.Actions == nil {
		return nil, errors.New("nil actions")
	}
	if action, ok := (*e.Spec.Strategy.Handlers.Actions)[name]; ok {
		return action, nil
	}
	return nil, errors.New("action with name " + name + " not found")
}

// Run tasks in a specified action.
func (e *Experiment) Run(name string) error {
	action, err := e.getAction(name)
	if err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), base.ContextKey("experiment"), e)
	for i := 0; i < len(*action); i++ {
		log.Info("------")
		err = (*action)[i].Run(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
