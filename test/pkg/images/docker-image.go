package images

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

type DockerImage struct {
	dockerClient *client.Client
	buildContext io.ReadCloser
	buildOptions types.ImageBuildOptions
}

func NewDockerImage(buildContextPath, dockerfilePath string) (
	i *DockerImage, e error,
) {
	i = &DockerImage{
		buildOptions: types.ImageBuildOptions{
			Tags:       make([]string, 0),
			Dockerfile: dockerfilePath,
			BuildArgs:  make(map[string]*string),
		},
	}

	i.dockerClient, e = client.NewClientWithOpts()
	if e != nil {
		return
	}

	i.buildContext, e = archive.TarWithOptions(buildContextPath,
		&archive.TarOptions{},
	)
	if e != nil {
		return
	}

	return
}

func (i *DockerImage) SetTag(tag string) {
	i.buildOptions.Tags = append(i.buildOptions.Tags, tag)

	return
}

func (i *DockerImage) SetBuildArg(key, value string) {
	i.buildOptions.BuildArgs[key] = &value

	return
}

func (i *DockerImage) Build(stream io.Writer) (e error) {
	var (
		response types.ImageBuildResponse
	)

	response, e = i.dockerClient.ImageBuild(
		context.Background(),
		i.buildContext,
		i.buildOptions,
	)
	if e != nil {
		return
	}

	defer response.Body.Close()

	_, e = io.Copy(stream, response.Body)
	if e != nil {
		return
	}

	return
}

func (i *DockerImage) Push(stream io.Writer) (e error) {
	const (
		registryAuth = "https://stackoverflow.com/questions/44400971/"
	)

	var (
		response io.ReadCloser

		j int
	)

	for j = 0; j < len(i.buildOptions.Tags); j++ {
		response, e = i.dockerClient.ImagePush(
			context.Background(),
			i.buildOptions.Tags[j],
			types.ImagePushOptions{
				RegistryAuth: registryAuth,
			},
		)
		if e != nil {
			return
		}

		_, e = io.Copy(stream, response)
		if e != nil {
			return
		}
	}

	return
}

func (i *DockerImage) Remove() (e error) {
	var (
		j int
	)

	for j = 0; j < len(i.buildOptions.Tags); j++ {
		_, e = i.dockerClient.ImageRemove(
			context.Background(),
			i.buildOptions.Tags[j],
			types.ImageRemoveOptions{},
		)
		if e != nil {
			return
		}
	}

	return
}
