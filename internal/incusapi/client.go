package incusapi

import "github.com/lxc/incus/client"

func NewClient() (incus.InstanceServer, error) {
	return incus.ConnectIncusUnix("", nil)
}
