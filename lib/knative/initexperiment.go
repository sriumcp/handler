package knative

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/iter8-tools/handler/base"
	"github.com/iter8-tools/handler/experiment"
	"k8s.io/apimachinery/pkg/types"

	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// TI struct contains a list of indexes (into knative.service.status.Traffic stanza) corresponding to the revisions used in an experiment.
type TI struct {
	Indexes []int `json:"indexes" yaml:"indexes"`
}

// InitExperimentTask enables initialization of knative experiments.
type InitExperimentTask struct {
	Library string `json:"library" yaml:"library"`
	Task    string `json:"task" yaml:"task"`
	With    TI     `json:"with,omitempty" yaml:"with,omitempty"`
}

// Run executes an InitExperimentTask
func (t *InitExperimentTask) Run(ctx context.Context) error {
	log.Trace("init experiment task run started...")
	var e *experiment.Experiment
	var err error
	if e, err = experiment.GetExperimentFromContext(ctx); err == nil {
		var nn *types.NamespacedName
		if nn, err = GetNamespacedNameForKsvc(e); err == nil {
			var ksvc *servingv1.Service
			log.Trace("Getting svc with namespaced name... ", *nn)
			if ksvc, err = GetKnativeSvc(nn); err == nil {
				if err = checkKsvcReadiness(ksvc); err == nil {
					if err = updateLocalExp(e, ksvc, t); err == nil {
						err = experiment.UpdateInClusterExperiment(e)
					}
				}
			}
		}
	}
	return err
}

// MakeInitExperiment converts an InitExperiment task spec into an InitExperimentTask.
func MakeInitExperiment(t *base.TaskSpec) (base.Task, error) {
	if t.Library != "knative" || t.Task != "init-experiment" {
		return nil, errors.New("library and task need to be 'knative' and 'init-experiment'")
	}
	var err error
	var jsonBytes []byte
	var it base.Task
	// convert t to jsonBytes
	jsonBytes, err = json.Marshal(t)
	// convert jsonString to ExecTask
	if err == nil {
		it = &InitExperimentTask{}
		err = json.Unmarshal(jsonBytes, &it)
	}
	return it, err
}
