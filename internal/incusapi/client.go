package incusapi

import (
	"os"

	"github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/cliconfig"
)

type Client incus.InstanceServer

func NewClient() (incus.InstanceServer, error) {
	path := os.ExpandEnv("$HOME/.config/incus/config.yml")
	conf, err := cliconfig.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	return conf.GetInstanceServer(conf.DefaultRemote)
}
