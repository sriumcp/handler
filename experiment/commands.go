package experiment

import (
	"errors"

	iter8 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/base"
)

// // extrapolate an experiment.
// func (e *Experiment) extrapolate() (er error) {
// 	if e == nil {
// 		return errors.New("extrapolate called on nil experiment")
// 	}
// 	if e.Spec.Strategy.Handlers == nil || e.Spec.Strategy.Handlers.Actions == nil {
// 		return nil
// 	}
// 	if rb, err := e.GetRecommendedBaseline(); err == nil {
// 		if version, err := e.getVersionDetail(rb); err == nil {
// 			if version.Tags == nil {
// 				return nil
// 			}
// 			for _, action := range *e.Spec.Strategy.Handlers.Actions {
// 				for i := 0; i < len(*action); i++ {
// 					if err = (*action)[i].Extrapolate(&base.Tags{M: version.Tags}); err != nil {
// 						log.Error("cannot extrapolate experiment: ", err)
// 						return err
// 					}
// 				}
// 			}
// 		} else {
// 			log.Error("error getting version detail")
// 			return err
// 		}
// 	} else {
// 		log.Error("error getting recommended baseline")
// 		return err
// 	}
// 	return nil
// }

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

// GetVersionDetail from the experiment for a named version.
func (e *Experiment) GetVersionDetail(versionName string) (*iter8.VersionDetail, error) {
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

// GetActionSpec gets a named action spec from an experiment.
// type ActionSpec []TaskSpec
func (e *Experiment) GetActionSpec(name string) ([]base.TaskSpec, error) {
	if e == nil {
		return nil, errors.New("GetActionSpec(...) called on nil experiment")
	}
	if e.Spec.Strategy.Handlers == nil {
		return nil, errors.New("nil handlers")
	}
	if e.Spec.Strategy.Handlers.Actions == nil {
		return nil, errors.New("nil actions")
	}
	if actionSpec, ok := (*e.Spec.Strategy.Handlers.Actions)[name]; ok {
		return actionSpec, nil
	}
	return nil, errors.New("action with name " + name + " not found")
}

// // Run extrapolates an experiment and runs a named action within it.
// func (e *Experiment) Run(name string) error {
// 	if e == nil {
// 		return errors.New("Run(...) called on nil experiment")
// 	}
// 	action, err := e.GetAction(name)
// 	if err != nil {
// 		return err
// 	}
// 	if err := e.extrapolate(); err != nil {
// 		log.Error(err)
// 		return err
// 	}
// 	ctx := context.WithValue(context.Background(), base.ContextKey("experiment"), e)
// 	for i := 0; i < len(*action); i++ {
// 		log.Info("------")
// 		err = (*action)[i].Run(ctx)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
