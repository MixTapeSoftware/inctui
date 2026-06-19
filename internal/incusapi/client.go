package incusapi

import "github.com/lxc/incus/client"

type Client incus.InstanceServer

func NewClient() (incus.InstanceServer, error) {
	return incus.ConnectIncusUnix("", nil)
}
