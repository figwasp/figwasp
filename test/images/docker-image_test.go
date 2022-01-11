package images

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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
	)

	var (
		createResponse container.ContainerCreateCreatedBody
		dockerClient   *client.Client
		e              error
		image          *dockerImage
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

	verifyHTTPResponse(t,
		fmt.Sprintf(addressFormat, hostIP, hostPort),
		path,
		message,
	)

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

func verifyHTTPResponse(t *testing.T, address, path, messageExpected string) {
	const (
		network   = "tcp"
		timeout   = time.Second
		urlFormat = "http://%s%s"
	)

	var (
		e                error
		httpResponse     *http.Response
		httpResponseBody []byte
		timer            context.Context
		url              string
	)

	timer, _ = context.WithTimeout(
		context.Background(),
		timeout,
	)

	for {
		_, e = net.Dial(network, address)
		if e == nil {
			break
		}

		if timer.Err() != nil {
			t.Error(e)

			break
		}
	}

	url = fmt.Sprintf(urlFormat,
		address,
		path,
	)

	httpResponse, e = http.Get(url)
	if e != nil {
		t.Error(e)
	}

	if httpResponse.StatusCode != http.StatusOK {
		t.Fail()
	}

	defer httpResponse.Body.Close()

	httpResponseBody, e = io.ReadAll(httpResponse.Body)
	if e != nil {
		t.Error(e)
	}

	if string(httpResponseBody) != messageExpected {
		t.Fail()
	}
}
