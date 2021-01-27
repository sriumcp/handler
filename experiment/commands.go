package experiment

import (
	"context"
	"errors"
	"fmt"

	iter8 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/base"
)

// extrapolate an experiment.
func (e *Experiment) extrapolate() (er error) {
	if e == nil {
		return errors.New("extrapolate called on nil experiment")
	}
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
					if err = (*action)[i].Extrapolate(&base.Tags{M: version.Tags}); err != nil {
						log.Error("cannot extrapolate experiment: ", err)
						return err
					}
				}
			}
		}
	} else {
		log.Error("error while getting recommended baseline")
		return err
	}
	return nil
}

// GetRecommendedBaseline from the experiment.
func (e *Experiment) GetRecommendedBaseline() (string, error) {
	if e == nil {
		return "", errors.New("GetRecommendedBaseline() called on nil experiment")
	}
	if e.Status.RecommendedBaseline == nil {
		return "", errors.New("Recommended baseline not found in experiment status")
	}
	return *e.Status.RecommendedBaseline, nil
}

// getVersionDetail from the experiment for a named version.
func (e *Experiment) getVersionDetail(versionName string) (*iter8.VersionDetail, error) {
	if e == nil {
		return nil, errors.New("getVersionDetail(...) called on nil experiment")
	}
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

// DryRun extrapolates an experiment and explains the tasks in the extrapolated experiment.
func (e *Experiment) DryRun() error {
	if e == nil {
		return errors.New("DryRun() called on nil experiment")
	}
	handlers := e.Spec.Strategy.Handlers
	if handlers == nil {
		fmt.Println("Experiment does not have a handler stanza")
		return nil
	}
	actions := handlers.Actions
	if actions == nil {
		fmt.Println("Experiment does not have a map of actions")
		return nil
	}
	if err := e.extrapolate(); err != nil {
		return err
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
		log.Trace("Action name: ", name)
		action.DryRun()
	}
	return nil
}

// DryRun explains the tasks in an action.
// Call this method on an action within an extrapolated experiment.
func (a *Action) DryRun() {
	if a == nil {
		panic("DryRun() called on nil action")
	}
	for i := 0; i < len(*a); i++ {
		task := (*a)[i]
		log.Trace("Action: ", *a)
		log.Trace("Task ptr: ", task)
		task.DryRun()
	}
}

// LocalRun extrapolates an experiment and runs the specified action or task locally.
func (e *Experiment) LocalRun(actionName string, task int) error {
	if e == nil {
		return errors.New("LocalRun(...) called on nil experiment")
	}
	handlers := e.Spec.Strategy.Handlers
	if handlers == nil || handlers.Actions == nil {
		return errors.New("Experiment does not have a handler stanza or actions")
	}
	action, err := e.getAction(actionName)
	if err != nil {
		return err
	}
	if err := e.extrapolate(); err != nil {
		return err
	}
	ctx := context.WithValue(context.Background(), base.ContextKey("experiment"), e)
	if task < 0 {
		return action.LocalRun(ctx)
	}
	if task >= 0 && task < len(*action) { // run task
		return (*action)[task].LocalRun(ctx)
	}
	return nil
}

// LocalRun runs the specified action locally.
// Call this method on an action within an extrapolated experiment.
func (a *Action) LocalRun(ctx context.Context) error {
	if a == nil {
		return errors.New("LocalRun(...) called on nil action")
	}
	var supported = true
	for i := 0; i < len(*a); i++ {
		supported = supported && (*a)[i].LocallyRunnable()
	}
	if !supported {
		return errors.New("local run called on action with tasks that are not locally runnable")
	}
	for i := 0; i < len(*a); i++ {
		if err := (*a)[i].LocalRun(ctx); err != nil {
			return err
		}
	}
	return nil
}

// getAction gets a named action from an experiment.
func (e *Experiment) getAction(name string) (*Action, error) {
	if e == nil {
		return nil, errors.New("getAction(...) called on nil experiment")
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

// Run extrapolates an experiment and runs a named action within it.
func (e *Experiment) Run(name string) error {
	if e == nil {
		return errors.New("Run(...) called on nil experiment")
	}
	action, err := e.getAction(name)
	if err != nil {
		return err
	}
	if err := e.extrapolate(); err != nil {
		log.Error(err)
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
