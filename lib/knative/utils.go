package knative

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iter8-tools/etc3/api/v2alpha1"
	"github.com/iter8-tools/handler/experiment"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
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
	log.Trace("Getting client")
	ksvc := &servingv1.Service{}
	log.Trace("Getting ksvc now...")
	if err = experiment.GetTypedObject(nn, ksvc); err == nil {
		log.Trace("got ksvc")
		return ksvc, nil
	}
	log.Error("cannot fetch ksvc from cluster")
	log.Error(err)
	return nil, err
}

// checkKsvcReadiness checks readiness of the knative service.
func checkKsvcReadiness(ksvc *servingv1.Service) error {
	var err = errors.New("knative service is not ready")
	ksvc1 := ksvc
	log.Trace("ksvc status...")
	log.Trace(ksvc.Status)
	for i := 0; i < experiment.NumAttempt; i++ {
		a, b, c := false, false, false
		log.Trace(ksvc1.Status.Conditions)
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
		// not ready yet
		time.Sleep(experiment.Period)
		ksvc1 = &servingv1.Service{}
		if err = experiment.GetTypedObject(&types.NamespacedName{
			Namespace: ksvc.Namespace,
			Name:      ksvc.Name,
		}, ksvc1); err != nil {
			log.Error(err)
			return errors.New("unable to get ksvc")
		}
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
	if tt, err = findConformanceTrafficTarget(e, ksvc, t); err == nil {
		// if a baseline is given, update revision
		if e.Spec.VersionInfo != nil {
			experiment.UpdateVariable(&e.Spec.VersionInfo.Baseline, "revision", tt.RevisionName)
		} else {
			// else attach new versionInfo for baseline
			e.Spec.VersionInfo = &v2alpha1.VersionInfo{
				Baseline: v2alpha1.VersionDetail{
					Name:      "baseline",
					Variables: []v2alpha1.Variable{{Name: "revision", Value: tt.RevisionName}},
				},
			}
		}
	}
	return err
}

// findConformanceTrafficTarget finds the traffic target from the given conformance experiment and ksvc.
func findConformanceTrafficTarget(e *experiment.Experiment, ksvc *servingv1.Service, t *InitExperimentTask) (*servingv1.TrafficTarget, error) {
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
	log.Trace("Task", *t)
	if len(t.With.Indexes) != 0 && len(t.With.Indexes) != 2 {
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

		var bw, cw int
		if bw, err = findTrafficTargetIndexInSpec(b, ksvc); err == nil {
			e.Spec.VersionInfo.Baseline.WeightObjRef = objRefFromFieldPath(bw, ksvc)
		} else {
			return err
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

		if cw, err = findTrafficTargetIndexInSpec(c, ksvc); err == nil {
			e.Spec.VersionInfo.Candidates[0].WeightObjRef = objRefFromFieldPath(cw, ksvc)
		} else {
			return err
		}
	}
	return err
}

// objRefFromFieldPath returns a weightObjRef to be inserted into the experiment using the given fieldPath fragment
func objRefFromFieldPath(fieldPathIndex int, ksvc *servingv1.Service) *corev1.ObjectReference {
	apiVersion, kind := ksvc.GetGroupVersionKind().ToAPIVersionAndKind()
	return &corev1.ObjectReference{
		Kind:       kind,
		Namespace:  ksvc.Namespace,
		Name:       ksvc.Name,
		APIVersion: apiVersion,
		FieldPath:  "/spec/traffic/" + fmt.Sprint(fieldPathIndex) + "/percent",
	}
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

// findTrafficTargetIndexInSpec finds the index of the given traffic target within the spec.traffic stanza of the given ksvc.
func findTrafficTargetIndexInSpec(tt *servingv1.TrafficTarget, ksvc *servingv1.Service) (int, error) {
	for i := 0; i < len(ksvc.Spec.Traffic); i++ {
		if ksvc.Spec.Traffic[i].RevisionName == tt.RevisionName {
			// matched by revisionName property
			return i, nil
		}
		if ksvc.Spec.Traffic[i].LatestRevision != nil {
			if *ksvc.Spec.Traffic[i].LatestRevision == true {
				if tt.LatestRevision != nil {
					if *tt.LatestRevision == true {
						// matched by latestRevision property
						return i, nil
					}
				}
			}
		}
	}
	return -1, errors.New("unable to find traffic target in spec")
}
