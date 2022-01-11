package images

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

type dockerImage struct {
	buildContext io.ReadCloser
	buildOptions types.ImageBuildOptions
}

func NewDockerImage(buildContextPath string) (i *dockerImage, e error) {
	i = new(dockerImage)

	i.buildContext, e = archive.TarWithOptions(buildContextPath,
		new(archive.TarOptions),
	)
	if e != nil {
		return
	}

	i.ClearTags()
	i.ClearBuildArgs()

	return
}

func (i *dockerImage) SetTag(tag string) {
	i.buildOptions.Tags = append(i.buildOptions.Tags, tag)

	return
}

func (i *dockerImage) ClearTags() {
	i.buildOptions.Tags = make([]string, 0)

	return
}

func (i *dockerImage) SetBuildArg(key, value string) {
	i.buildOptions.BuildArgs[key] = &value

	return
}

func (i *dockerImage) ClearBuildArgs() {
	i.buildOptions.BuildArgs = make(map[string]*string)

	return
}

func (i *dockerImage) Build(dockerClient *client.Client, output io.Writer) (
	e error,
) {
	var (
		buildResponse types.ImageBuildResponse
	)

	buildResponse, e = dockerClient.ImageBuild(
		context.Background(),
		i.buildContext,
		i.buildOptions,
	)
	if e != nil {
		return
	}

	defer buildResponse.Body.Close()

	_, e = io.Copy(output, buildResponse.Body)
	if e != nil {
		return
	}

	return
}
