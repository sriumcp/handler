package experiment

import (
	"context"
	"errors"

	iter8 "github.com/iter8-tools/etc3/api/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Much of this k8sclient code is based on the following tutorial:
// https://itnext.io/how-to-generate-client-codes-for-kubernetes-custom-resource-definitions-crd-b4b9907769ba

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return iter8.GroupVersion.WithResource(resource).GroupResource()
}

// GetClient constructs and returns a K8s client with using the rest config.
// The returned client has experiment.Experiment type registered.
func GetClient(restConf *rest.Config) (rc client.Client, err error) {
	scheme := runtime.NewScheme()
	var addKnownTypes = func(s *runtime.Scheme) error {
		s.AddKnownTypes(iter8.GroupVersion, &Experiment{})
		scheme.AddKnownTypes(
			iter8.GroupVersion,
			&metav1.Status{},
		)
		metav1.AddToGroupVersion(scheme, iter8.GroupVersion)
		return nil
	}
	// runtime.NewSchemeBuilder appears to be a wrapper around addKnownTypes
	// the latter does not return errors, the former does
	var schemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	err = schemeBuilder.AddToScheme(scheme)

	if err == nil {
		rc, err = client.New(restConf, client.Options{
			Scheme: scheme,
		})
		if err == nil {
			return rc, nil
		}
	}
	return nil, errors.New("cannot get client using rest config")
}

// FromCluster fetches an experiment from k8s cluster.
func (b *Builder) FromCluster(name string, namespace string, restClient client.Client) *Builder {
	// get the exp; this is a handler (enhanced) exp -- not just an iter8 exp.
	exp := &Experiment{
		Experiment: *iter8.NewExperiment(name, namespace).Build(),
	}
	var err error
	if err = restClient.Get(context.Background(), client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, exp); err == nil {
		b.exp = exp
		return b
	}
	log.Error(err)
	b.err = errors.New("cannot build experiment from cluster")
	return b
}
