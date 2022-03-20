package images

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

type DockerImage struct {
	dockerClient *client.Client
	buildContext io.ReadCloser
	buildOptions types.ImageBuildOptions
	outputStream io.Writer
}

func NewDockerImage(
	buildContextPath, dockerfilePath string, options ...dockerImageOption,
) (
	i *DockerImage, e error,
) {
	var (
		option dockerImageOption
	)

	i = &DockerImage{
		buildOptions: types.ImageBuildOptions{
			Tags:       make([]string, 0),
			Dockerfile: dockerfilePath,
			BuildArgs:  make(map[string]*string),
		},
		outputStream: io.Discard,
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

	for _, option = range options {
		e = option(i)
		if e != nil {
			return
		}
	}

	e = i.build()
	if e != nil {
		return
	}

	return
}

func (i *DockerImage) build() (e error) {
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

	_, e = io.Copy(i.outputStream, response.Body)
	if e != nil {
		return
	}

	return
}

func (i *DockerImage) Push(stream io.Writer) error {
	const (
		registryAuth = "https://stackoverflow.com/questions/44400971/"
	)

	return i.push(stream, registryAuth)
}

func (i *DockerImage) PushWithBasicAuth(
	stream io.Writer, username, password string,
) (
	e error,
) {
	var (
		jsonStruct struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		jsonBytes    []byte
		registryAuth string
	)

	jsonStruct.Username = username
	jsonStruct.Password = password

	jsonBytes, e = json.Marshal(jsonStruct)
	if e != nil {
		return
	}

	registryAuth = base64.StdEncoding.EncodeToString(jsonBytes)

	e = i.push(stream, registryAuth)
	if e != nil {
		return
	}

	return
}

func (i *DockerImage) push(stream io.Writer, registryAuth string) (e error) {
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

func (i *DockerImage) Destroy() (e error) {
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

type dockerImageOption func(*DockerImage) error

func WithTag(tag string) (option dockerImageOption) {
	option = func(i *DockerImage) (e error) {
		i.buildOptions.Tags = append(i.buildOptions.Tags, tag)

		return
	}

	return
}

func WithBuildArg(key, value string) (option dockerImageOption) {
	option = func(i *DockerImage) (e error) {
		i.buildOptions.BuildArgs[key] = &value

		return
	}

	return
}

func WithOutputStream(writer io.Writer) (option dockerImageOption) {
	option = func(i *DockerImage) (e error) {
		i.outputStream = writer

		return
	}

	return
}
