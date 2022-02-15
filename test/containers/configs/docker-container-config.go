package configs

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

type dockerContainerConfig struct {
	config     *container.Config
	hostConfig *container.HostConfig
}

func NewDockerContainerConfig(imageRef string) (
	c *dockerContainerConfig,
) {
	const (
		attachStreams = false
	)

	c = &dockerContainerConfig{
		config: &container.Config{
			ExposedPorts: nat.PortSet{},
			Tty:          attachStreams,
			Image:        imageRef,
		},
		hostConfig: &container.HostConfig{
			PortBindings: nat.PortMap{},
		},
	}

	return
}

func (c *dockerContainerConfig) Config() *container.Config {
	return c.config
}

func (c *dockerContainerConfig) HostConfig() *container.HostConfig {
	return c.hostConfig
}

func (c *dockerContainerConfig) PublishTCPPort(port, hostIP, hostPort string) {
	const (
		natPortFormat = "%s/tcp"
	)

	var (
		natPort nat.Port
	)

	natPort = nat.Port(
		fmt.Sprintf(natPortFormat, port),
	)

	c.config.ExposedPorts[natPort] = struct{}{}

	c.hostConfig.PortBindings[natPort] = append(
		c.hostConfig.PortBindings[natPort],
		nat.PortBinding{
			HostIP:   hostIP,
			HostPort: hostPort,
		},
	)

	return
}
