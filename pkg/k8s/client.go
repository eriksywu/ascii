package k8s

import (
	"k8s.io/api/core/v1"
)

type K8sClient interface {
	GetServiceEndpoints(namespace, service string) (*v1.Endpoints, error)
}
