package incusapi

import (
	"github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
	"sort"
)

type InstanceServer struct {
	server incus.InstanceServer
}

func NewInstanceServer() (*InstanceServer, error) {
	server, err := NewClient()
	return &InstanceServer{server: server}, err
}

// Get a list of incus instances
func (f *InstanceServer) Instances() ([]api.Instance, error) {
	instances, err := f.server.GetInstances(api.InstanceTypeAny)
	if err != nil {
		return nil, err
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})
	return instances, nil
}

func (f *InstanceServer) Instance(name string) (*api.Instance, error) {
	instance, _, err := f.server.GetInstance(name)
	return instance, err
}

// Get the live state of an incus instance (e.g. CPU usage)
func (f *InstanceServer) InstanceState(name string) (*api.InstanceState, error) {
	// Ignore the ETAG for now - used in versioning the resource for optimistic conflict resolution
	// we don't need it.
	instanceState, _, err := f.server.GetInstanceState(name)
	return instanceState, err
}

func (f *InstanceServer) ToggleInstance(name string, statusCode api.StatusCode) api.StatusCode {
	var action string
	var transitionalStatusCode api.StatusCode
	switch statusCode {
	case api.Running:
		action = "stop"
		transitionalStatusCode = api.Stopping
	case api.Stopped:
		action = "start"
		transitionalStatusCode = api.Starting
	}

	if action != "" {
		statePut := api.InstanceStatePut{Action: action}
		f.server.UpdateInstanceState(name, statePut, "")
		return transitionalStatusCode
	}
	return statusCode
}
