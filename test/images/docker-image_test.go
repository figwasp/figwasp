package images

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/joel-ling/alduin/test/httpclients"
	"github.com/stretchr/testify/assert"
)

func TestDockerImage(t *testing.T) {
	const (
		addressFormat    = "%s:%s"
		attachStreams    = false
		buildArgKey      = "APP_MESSAGE"
		buildContextPath = "context"
		containerName    = "test"
		containerPort    = "8000/tcp"
		hostIP           = "127.0.0.1"
		hostPort         = "8000"
		imageTag         = "test:v0"
		message          = "Hello, World!"
		path             = "/"
		timeout          = time.Second
		urlFormat        = "http://%s:%s"
	)

	type (
		httpResponseVerifier interface {
			OK(chan<- bool)
		}
	)

	var (
		createResponse container.ContainerCreateCreatedBody
		dockerClient   *client.Client
		e              error
		image          *dockerImage
		ok             chan bool
		verifier       httpResponseVerifier
	)

	image, e = NewDockerImage(buildContextPath)
	if e != nil {
		t.Error(e)
	}

	image.SetTag(imageTag)

	image.SetBuildArg(buildArgKey, message)

	dockerClient, e = client.NewClientWithOpts()
	if e != nil {
		t.Error(e)
	}

	e = image.Build(dockerClient, os.Stdout)
	if e != nil {
		t.Error(e)
	}

	createResponse, e = dockerClient.ContainerCreate(
		context.Background(),
		&container.Config{
			ExposedPorts: nat.PortSet{
				containerPort: struct{}{},
			},
			Tty:   attachStreams,
			Image: imageTag,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				containerPort: []nat.PortBinding{
					{
						HostIP:   hostIP,
						HostPort: hostPort,
					},
				},
			},
		},
		nil,
		nil,
		containerName,
	)
	if e != nil {
		t.Error(e)
	}

	e = dockerClient.ContainerStart(
		context.Background(),
		createResponse.ID,
		types.ContainerStartOptions{},
	)
	if e != nil {
		t.Error(e)
	}

	verifier = httpclients.NewSimpleHTTPClient(
		fmt.Sprintf(urlFormat, hostIP, hostPort),
		message,
		timeout,
	)

	ok = make(chan bool)

	go verifier.OK(ok)

	assert.True(t, <-ok)

	e = dockerClient.ContainerStop(
		context.Background(),
		createResponse.ID,
		nil,
	)
	if e != nil {
		t.Error(e)
	}

	e = dockerClient.ContainerRemove(
		context.Background(),
		createResponse.ID,
		types.ContainerRemoveOptions{},
	)
	if e != nil {
		t.Error(e)
	}
}
