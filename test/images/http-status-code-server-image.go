package images

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/joel-ling/alduin/test/constants"
)

type httpStatusCodeServerImage struct {
	dockerClient *client.Client
	buildContext io.ReadCloser
	buildOptions types.ImageBuildOptions
}

func NewHTTPStatusCodeServerImage(statusCode int, buildContextPath string) (
	i *httpStatusCodeServerImage, e error,
) {
	const (
		buildArgumentKey = "STATUS_CODE"
	)

	i = &httpStatusCodeServerImage{
		buildOptions: types.ImageBuildOptions{
			Tags:      make([]string, 1),
			BuildArgs: make(map[string]*string),
		},
	}

	i.dockerClient, e = client.NewClientWithOpts()
	if e != nil {
		return
	}

	i.setTag(constants.StatusCodeServerImageRef)

	i.setBuildArg(buildArgumentKey,
		fmt.Sprint(statusCode),
	)

	i.buildContext, e = archive.TarWithOptions(buildContextPath,
		new(archive.TarOptions),
	)
	if e != nil {
		return
	}

	return
}

func (i *httpStatusCodeServerImage) setTag(tag string) {
	i.buildOptions.Tags[0] = tag
}

func (i *httpStatusCodeServerImage) setBuildArg(key, value string) {
	i.buildOptions.BuildArgs[key] = &value
}

func (i *httpStatusCodeServerImage) Build() (e error) {
	var (
		buildResponse types.ImageBuildResponse
	)

	buildResponse, e = i.dockerClient.ImageBuild(
		context.Background(),
		i.buildContext,
		i.buildOptions,
	)
	if e != nil {
		return
	}

	_, e = io.Copy(os.Stderr, buildResponse.Body)
	if e != nil {
		return
	}

	return
}

func (i *httpStatusCodeServerImage) Remove() (e error) {
	_, e = i.dockerClient.ImageRemove(
		context.Background(),
		i.buildOptions.Tags[0],
		types.ImageRemoveOptions{},
	)
	if e != nil {
		return
	}

	return
}
