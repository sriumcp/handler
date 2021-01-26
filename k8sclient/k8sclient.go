// Package k8sclient enables in-cluster interaction with Kubernetes API server.
package k8sclient

import (
	"errors"

	etc3 "github.com/iter8-tools/etc3/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// GetInClusterClient constructs and returns an in-cluster K8s client.
func GetInClusterClient() (rc client.Client, err error) {
	crScheme := runtime.NewScheme()
	err = etc3.AddToScheme(crScheme)
	if err == nil {
		var restConf *rest.Config
		restConf, err = config.GetConfig()
		if err == nil {
			rc, err = client.New(restConf, client.Options{
				Scheme: crScheme,
			})
			if err == nil {
				return rc, nil
			}
		}
	}
	return nil, errors.New("cannot construct in-cluster client")
}
