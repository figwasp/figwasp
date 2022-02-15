package configs

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

type DockerContainerConfig struct {
	config     *container.Config
	hostConfig *container.HostConfig
}

func NewDockerContainerConfig(imageRef string) (
	c *DockerContainerConfig, e error,
) {
	const (
		attachStreams = false
	)

	c = &DockerContainerConfig{
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

func (c *DockerContainerConfig) Config() *container.Config {
	return c.config
}

func (c *DockerContainerConfig) HostConfig() *container.HostConfig {
	return c.hostConfig
}

func (c *DockerContainerConfig) PublishTCPPort(port, hostIP, hostPort string) {
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
