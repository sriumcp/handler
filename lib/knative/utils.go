package knative

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/experiment"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNamespacedNameForKsvc parses the target of the experiment and returns name and namespace of the ksvc.
func GetNamespacedNameForKsvc(e *experiment.Experiment) (*types.NamespacedName, error) {
	if nn := strings.Split(e.Spec.Target, "/"); len(nn) == 2 {
		return &types.NamespacedName{
			Namespace: nn[0],
			Name:      nn[1],
		}, nil
	}
	return nil, errors.New("unable to parse target into name and namespace of the ksvc")
}

// GetKnativeSvc fetches the Knative service given its name and namespace.
func GetKnativeSvc(nn *types.NamespacedName) (*servingv1.Service, error) {
	var err error
	var rc client.Client
	if rc, err = experiment.GetClient(); err == nil {
		ksvc := &servingv1.Service{}
		if err = rc.Get(context.Background(), client.ObjectKey{
			Namespace: nn.Namespace,
			Name:      nn.Name,
		}, ksvc); err == nil {
			return ksvc, nil
		}
	}
	return nil, err
}

// checkKsvcReadiness checks readiness of the knative service.
func checkKsvcReadiness(ksvc *servingv1.Service) error {
	var err = errors.New("knative service is not ready")
	ksvc1 := ksvc.DeepCopy()
	for i := 0; i < experiment.NumAttempt; i++ {
		err = experiment.GetTypedObject(&types.NamespacedName{
			Name:      ksvc1.Name,
			Namespace: ksvc1.Namespace,
		}, ksvc1)
		a, b, c := false, false, false
		for _, condition := range ksvc1.Status.Conditions {
			if condition.Type == servingv1.ServiceConditionReady &&
				condition.Status == corev1.ConditionTrue {
				a = true
			}
			if condition.Type == servingv1.ServiceConditionConfigurationsReady &&
				condition.Status == corev1.ConditionTrue {
				b = true
			}
			if condition.Type == servingv1.ServiceConditionRoutesReady &&
				condition.Status == corev1.ConditionTrue {
				c = true
			}
		}
		// everything looks good
		if a && b && c {
			return nil
		}
		time.Sleep(experiment.Period)
	}
	return err
}

// updateLocalExp updates the given Knative experiment struct.
func updateLocalExp(e *experiment.Experiment, ksvc *servingv1.Service, t *InitExperimentTask) error {
	switch e.Spec.Strategy.TestingPattern {
	case v2alpha1.TestingPatternConformance:
		return updateLocalConformanceExp(e, ksvc, t)
	case v2alpha1.TestingPatternCanary:
		return updateLocalCanaryExp(e, ksvc, t)
	default:
		return errors.New("unsupported testing pattern found in experiment")
	}
}

// updateLocalConformanceExp updates the given Knative conformance experiment struct.
func updateLocalConformanceExp(e *experiment.Experiment, ksvc *servingv1.Service, t *InitExperimentTask) error {
	var err error
	if len(t.With.Indexes) > 1 {
		return errors.New("performance experiment cannot have more than one traffic target")
	}
	var tt *servingv1.TrafficTarget
	if tt, err = findPerformanceTrafficTarget(e, ksvc, t); err == nil {
		// if a baseline is given, update revision
		if e.Spec.VersionInfo != nil {
			experiment.UpdateVariable(&e.Spec.VersionInfo.Baseline, "revision", tt.RevisionName)
		} else {
			// else attach new versionInfo for baseline
			e.Spec.VersionInfo = &v2alpha1.VersionInfo{
				Baseline: v2alpha1.VersionDetail{
					Name: "baseline",
					Variables: []v2alpha1.Variable{{
						Name:  "revision",
						Value: tt.RevisionName,
					}},
				},
			}
		}
	}
	return err
}

// findPerformanceTrafficTarget finds the traffic target from the given performance experiment and ksvc.
func findPerformanceTrafficTarget(e *experiment.Experiment, ksvc *servingv1.Service, t *InitExperimentTask) (*servingv1.TrafficTarget, error) {
	numTargets := len(ksvc.Status.Traffic)
	var targetIndex int
	if len(t.With.Indexes) == 0 {
		targetIndex = numTargets - 1
	} else {
		targetIndex = t.With.Indexes[0]
	}
	if targetIndex >= numTargets {
		return nil, errors.New("invalid target index specified in initExperimentTask")
	}
	return &ksvc.Status.Traffic[targetIndex], nil
}

// updateLocalCanaryExp updates the given Knative canary experiment struct.
func updateLocalCanaryExp(e *experiment.Experiment, ksvc *servingv1.Service, t *InitExperimentTask) error {
	var err error
	if len(t.With.Indexes) != 0 || len(t.With.Indexes) != 2 {
		return errors.New("invalid number of targets in canary experiment: " + fmt.Sprint(len(t.With.Indexes)))
	}
	var b, c *servingv1.TrafficTarget
	if b, c, err = findCanaryTrafficTargets(e, ksvc, t); err == nil {
		// fix baseline
		if e.Spec.VersionInfo != nil {
			experiment.UpdateVariable(&e.Spec.VersionInfo.Baseline, "revision", b.RevisionName)
		} else {
			// else attach new versionInfo with baseline
			e.Spec.VersionInfo = &v2alpha1.VersionInfo{
				Baseline: v2alpha1.VersionDetail{
					Name: "baseline",
					Variables: []v2alpha1.Variable{
						{
							Name:  "revision",
							Value: b.RevisionName,
						},
					},
				},
			}
		}

		// fix candidate
		if len(e.Spec.VersionInfo.Candidates) == 0 {
			e.Spec.VersionInfo.Candidates = []v2alpha1.VersionDetail{
				{
					Name: "candidate",
					Variables: []v2alpha1.Variable{
						{
							Name:  "revision",
							Value: c.RevisionName,
						},
					},
				},
			}
		} else {
			experiment.UpdateVariable(&e.Spec.VersionInfo.Candidates[0], "revision", c.RevisionName)
		}
	}
	return err
}

// findCanaryTrafficTargets finds the traffic targets from the given canary experiment and ksvc.
func findCanaryTrafficTargets(e *experiment.Experiment, ksvc *servingv1.Service, t *InitExperimentTask) (*servingv1.TrafficTarget, *servingv1.TrafficTarget, error) {
	numTargets := len(ksvc.Status.Traffic)
	if numTargets < 2 {
		return nil, nil, errors.New("insufficient number of traffic targets in knative service")
	}
	baselineIndex := numTargets - 2
	candidateIndex := numTargets - 1

	switch len(t.With.Indexes) {
	case 0:
		// do nothing
	case 2:
		baselineIndex = t.With.Indexes[0]
		candidateIndex = t.With.Indexes[1]
	default:
		return nil, nil, errors.New("invalid number of traffic target indexes for canary experiment: " + fmt.Sprint(len(t.With.Indexes)))
	}

	if baselineIndex < 0 || candidateIndex < 0 || baselineIndex > numTargets-1 || candidateIndex > numTargets-1 {
		return nil, nil, errors.New("baseline or candidate index is out of bounds")
	}

	return &ksvc.Status.Traffic[baselineIndex], &ksvc.Status.Traffic[candidateIndex], nil
}
