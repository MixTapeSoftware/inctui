package incusapi

import (
	"github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
	"sort"
)

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
