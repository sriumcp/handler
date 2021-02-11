// Package experiment enables construction of an experiment object with handler/task lists within it.
package experiment

import (
	"context"
	"errors"

	iter8 "github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/base"
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

// ActionMap type represents a map whose keys are actions names, and whose values are slices of TaskSpecs.
type ActionMap map[string][]base.TaskSpec

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

// GetExperimentFromContext gets the experiment object from given context.
func GetExperimentFromContext(ctx context.Context) (*Experiment, error) {
	//	ctx := context.WithValue(context.Background(), base.ContextKey("experiment"), e)
	if v := ctx.Value("experiment"); v != nil {
		log.Debug("found experiment")
		var e *Experiment
		var ok bool
		if e, ok = v.(*Experiment); !ok {
			return nil, errors.New("context has experiment value with wrong type")
		}
		return e, nil
	}
	return nil, errors.New("context has no experiment key")
}

// Extrapolate extrapolates input arguments based on tags of the recommended baseline in the experiment.
func (exp *Experiment) Extrapolate(inputArgs []string) ([]string, error) {
	var recommendedBaseline string
	var args []string
	var err error
	if recommendedBaseline, err = exp.GetRecommendedBaseline(); err == nil {
		var versionDetail *iter8.VersionDetail
		if versionDetail, err = exp.GetVersionDetail(recommendedBaseline); err == nil {
			// get the tags
			tags := base.Tags{M: make(map[string]string)}
			tags.M["name"] = versionDetail.Name
			for i := 0; i < len(versionDetail.Variables); i++ {
				tags.M[versionDetail.Variables[i].Name] = tags.M[versionDetail.Variables[i].Value]
			}
			args := make([]string, len(inputArgs))
			for i := 0; i < len(args); i++ {
				if args[i], err = tags.Extrapolate(&inputArgs[i]); err != nil {
					break
				}
			}
		}
	}
	return args, err
}
