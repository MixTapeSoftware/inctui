package incusapi

import (
	"github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
	"sort"
)

type InstanceFetcher struct {
	server incus.InstanceServer
}

func NewInstanceFetcher() (*InstanceFetcher, error) {
	server, err := NewClient()
	return &InstanceFetcher{server: server}, err
}

// Get a list of incus instances
func (f *InstanceFetcher) Instances() ([]api.Instance, error) {
	instances, err := f.server.GetInstances(api.InstanceTypeAny)
	if err != nil {
		return nil, err
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})
	return instances, nil
}

// Get the live state of an incus instance (e.g. CPU usage)
func (f *InstanceFetcher) InstanceState(name string) (*api.InstanceState, error) {
	// Ignore the ETAG for now - used in versioning the resource for optimistic conflict resolution
	// we don't need it.
	instanceState, _, err := f.server.GetInstanceState(name)
	return instanceState, err
}
