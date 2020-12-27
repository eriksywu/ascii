package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/eriksywu/ascii/pkg/k8s"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
)

type WatchOpts struct {
	namespace string
	service string
}

// PoorMansProxy is as its name suggests - a poor man's L3+L7 proxy/loadbalancer for the ascii service
// In a pre-prod env, assume we want to deploy ascii as a service on a public cloud.
// Great - but using a LoadBalancer is expensive and unnecessary.
// Also, a client app may want to get a list of all images managed by all instances of ascii.
// Barring more work to redesign the system, the only way to accomplish that is to do an explicit list call for each instance.
// We can 1) deploy ascii as a headless service and have the client code do a manual dns lookup for all A records for the ascii service.
// This can get tricky if the client code isn't running in-cluster and part of the pod overlay network (if using kubenet)
// or 2) deploy this poor man's proxy to do that for you. This implementation has the added benefit of being k8s-aware thus bypassing
// the need to do any direct DNS lookups. We can in fact deploy the ascii service as a regular CLusterIP service instead of a headless service.
// This can also be used in a prod env if we use an Ingress+IngressController instead of directly exposing the ascii service via LoadBalancer

type PoorMansProxy struct {
	k8sClient           k8s.K8sClient
	WatchOpts			*WatchOpts
}

type endpointPair struct {
	ip string
	port int32
}

func (ep endpointPair) Address() string {
	return fmt.Sprintf("http://%s:%d", ep.ip, ep.port)
}

func (p *PoorMansProxy) ListASCIIImages() ([]uuid.UUID, error) {
	endpoints, err := p.getAllServiceEndpoints()
	if err != nil {
		return nil, err
	}
	idList := make([]uuid.UUID, 0)
	for _, ep := range endpoints {
		listResponse, err := http.Get(ep.Address()+"/images")
		var ids []uuid.UUID

		body, err := ioutil.ReadAll(listResponse.Body)
		listResponse.Body.Close()
		if err != nil {
			return nil, err
		}
		json.Unmarshal(body, ids)
		idList = append(idList, ids...)
	}
	return idList, nil
}

// TODO do some caching or something here
func (p *PoorMansProxy) GetASCIIImage(id uuid.UUID) (bool, string, error) {
	endpoints, err := p.getAllServiceEndpoints()
	if err != nil {
		return false, "", err
	}
	for _, ep := range endpoints{
		response, err := http.Get(fmt.Sprintf("%s/images/%s", ep.Address(), id.String()))
		if err != nil {
			return false, "", err
		}
		if response.StatusCode != http.StatusOK {
			continue
		}
		body, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return false, "", err
		}
		return true, string(body), nil
	}
	return false, "", fmt.Errorf("not found etc etc todo add some actual errors here or something")
}

func (p *PoorMansProxy) PushASCIIImage(asciiImage string, id uuid.UUID) error {
	//no-op
	//NOTE: refactor
	return nil
}

func (p *PoorMansProxy) getAllServiceEndpoints() ([]endpointPair, error) {
	endpoints, err := p.k8sClient.GetServiceEndpoints(p.WatchOpts.namespace, p.WatchOpts.service)
	if err != nil {
		return nil, err
	}
	endpointPairs := make([]endpointPair, 0, len(endpoints.Subsets))
	for _, ep := range endpoints.Subsets {
		for _, addr := range ep.Addresses {
			for _, p := range ep.Ports {
				endpointPairs = append(endpointPairs, endpointPair{ip: addr.IP, port: p.Port})
			}
		}
	}
	return endpointPairs, nil
}
