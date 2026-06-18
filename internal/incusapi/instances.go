package incusapi

import (
	"github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
	"sort"
)

// Get a list of incus instances
func Instances(client incus.InstanceServer) ([]api.Instance, error) {
	instances, err := client.GetInstances(api.InstanceTypeAny)
	if err != nil {
		return nil, err
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Name < instances[j].Name
	})
	return instances, nil
}

// Get the live state of an incus instance (e.g. CPU usage)
func InstanceState(client incus.InstanceServer, name string) (*api.InstanceState, error) {
	// Ignore the ETAG for now - used in versioning the resource for optimistic conflict resolution
	// we don't need it.
	instanceState, _, err := client.GetInstanceState(name)
	return instanceState, err
}
