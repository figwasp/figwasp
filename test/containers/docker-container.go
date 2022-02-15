package containers

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/joel-ling/alduin/test/constants"
	"github.com/joel-ling/alduin/test/containers/configs"
)

type dockerContainer struct {
	dockerClient *client.Client
	metadata     container.ContainerCreateCreatedBody
}

func NewDockerContainer(imageRef, containerPort string) (
	c *dockerContainer, e error,
) {
	var (
		consoleOutput io.ReadCloser

		config interface {
			Config() *container.Config
			HostConfig() *container.HostConfig
			PublishTCPPort(string, string, string)
		}
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

	config = configs.NewDockerContainerConfig(imageRef)

	config.PublishTCPPort(containerPort,
		constants.StatusCodeServerIP,
		constants.StatusCodeServerPort,
	)

	c.metadata, e = c.dockerClient.ContainerCreate(
		context.Background(),
		config.Config(),
		config.HostConfig(),
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
