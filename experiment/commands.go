package experiment

import (
	"errors"

	"github.com/iter8-tools/etc3/api/v2alpha1"
	iter8 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/base"
)

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
	if e.Spec.Strategy.Actions == nil {
		return nil, errors.New("nil actions")
	}
	if actionSpec, ok := e.Spec.Strategy.Actions[name]; ok {
		return actionSpec, nil
	}
	return nil, errors.New("action with name " + name + " not found")
}

// UpdateVariable updates a variable within the given VersionDetail. If the variable is already present in the VersionDetail object, the pre-existing value takes precedence and is retained; if not, the new value is inserted.
func UpdateVariable(v *v2alpha1.VersionDetail, name string, value string) error {
	if v == nil {
		return errors.New("nil valued version detail")
	}
	for i := 0; i < len(v.Variables); i++ {
		if v.Variables[i].Name == name {
			log.Info("variable with same name already present in version detail")
			return nil
		}
	}
	v.Variables = append(v.Variables, v2alpha1.Variable{
		Name:  name,
		Value: value,
	})
	return nil
}
