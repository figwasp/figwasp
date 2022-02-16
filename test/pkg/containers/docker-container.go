package containers

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/joel-ling/alduin/pkg/containers/configs"
)

type DockerContainer struct {
	dockerClient *client.Client
	containerID  string
}

func NewDockerContainer(
	imageRef, name string, config *configs.DockerContainerConfig,
	stream io.Writer,
) (
	c *DockerContainer, e error,
) {
	var (
		metadata container.ContainerCreateCreatedBody
		response io.ReadCloser
	)

	c = &DockerContainer{}

	c.dockerClient, e = client.NewClientWithOpts()
	if e != nil {
		return
	}

	response, e = c.dockerClient.ImagePull(
		context.Background(),
		imageRef,
		types.ImagePullOptions{},
	)
	if e != nil {
		return
	}

	_, e = io.Copy(stream, response)
	if e != nil {
		return
	}

	metadata, e = c.dockerClient.ContainerCreate(
		context.Background(),
		config.Config(),
		config.HostConfig(),
		nil,
		nil,
		name,
	)
	if e != nil {
		return
	}

	c.containerID = metadata.ID

	e = c.dockerClient.ContainerStart(
		context.Background(),
		c.containerID,
		types.ContainerStartOptions{},
	)
	if e != nil {
		return
	}

	return
}

func (c *DockerContainer) Remove() (e error) {
	e = c.dockerClient.ContainerStop(
		context.Background(),
		c.containerID,
		nil,
	)
	if e != nil {
		return
	}

	e = c.dockerClient.ContainerRemove(
		context.Background(),
		c.containerID,
		types.ContainerRemoveOptions{},
	)
	if e != nil {
		return
	}

	return
}
