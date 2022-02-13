package containers

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/joel-ling/alduin/test/constants"
)

type dockerContainer struct {
	dockerClient *client.Client
	metadata     container.ContainerCreateCreatedBody
}

func NewDockerContainer(imageRef, containerPort string) (
	c *dockerContainer, e error,
) {
	const (
		attachStreams = false
	)

	var (
		consoleOutput io.ReadCloser
	)

	c = &dockerContainer{}

	c.dockerClient, e = client.NewClientWithOpts()
	if e != nil {
		return
	}

	consoleOutput, e = c.dockerClient.ImagePull(
		context.Background(),
		imageRef,
		types.ImagePullOptions{},
	)
	if e != nil {
		return
	}

	io.Copy(os.Stderr, consoleOutput)

	c.metadata, e = c.dockerClient.ContainerCreate(
		context.Background(),
		&container.Config{
			ExposedPorts: nat.PortSet{
				nat.Port(containerPort): struct{}{},
			},
			Tty:   attachStreams,
			Image: imageRef,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(containerPort): []nat.PortBinding{
					{
						HostIP:   constants.StatusCodeServerIP,
						HostPort: constants.StatusCodeServerPort,
					},
				},
			},
		},
		nil,
		nil,
		constants.StatusCodeServerContainerName,
	)
	if e != nil {
		return
	}

	e = c.dockerClient.ContainerStart(
		context.Background(),
		c.metadata.ID,
		types.ContainerStartOptions{},
	)
	if e != nil {
		return
	}

	return
}

func (c *dockerContainer) Destroy() (e error) {
	e = c.dockerClient.ContainerStop(
		context.Background(),
		c.metadata.ID,
		nil,
	)
	if e != nil {
		return
	}

	e = c.dockerClient.ContainerRemove(
		context.Background(),
		c.metadata.ID,
		types.ContainerRemoveOptions{},
	)
	if e != nil {
		return
	}

	return
}
