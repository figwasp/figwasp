package test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joel-ling/alduin/test/pkg/clients"
	"github.com/joel-ling/alduin/test/pkg/containers"
	"github.com/joel-ling/alduin/test/pkg/containers/configs"
	"github.com/joel-ling/alduin/test/pkg/images"
	"github.com/joel-ling/alduin/test/pkg/repositories"
)

func TestTest(t *testing.T) {
	const (
		localhost = "127.0.0.1"

		repositoryPort = 5000

		buildContextPath = "../.."
		dockerfilePath   = "test/build/http-status-code-server/Dockerfile"
		imageName        = "http-status-code-server"
		imageRefFormat   = "%s/%s"

		serverPortKey   = "SERVER_PORT"
		serverPortValue = 8000
		statusCodeKey   = "STATUS_CODE"
		statusCodeValue = http.StatusTeapot

		scheme  = "http"
		timeout = time.Second
	)

	var (
		repository        *repositories.DockerRegistry
		repositoryAddress net.TCPAddr

		image    *images.DockerImage
		imageRef string

		config    *configs.DockerContainerConfig
		container *containers.DockerContainer

		client        *clients.HTTPClient
		endpoint      url.URL
		serverAddress net.TCPAddr

		status int

		e error
	)

	// set up container image repository

	repositoryAddress = net.TCPAddr{
		IP:   net.ParseIP(localhost),
		Port: repositoryPort,
	}

	repository, e = repositories.NewDockerRegistry(repositoryAddress)
	if e != nil {
		t.Error(e)
	}

	defer repository.Close()

	// build and push container image of server to repository

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath)
	if e != nil {
		t.Error(e)
	}

	imageRef = fmt.Sprintf(imageRefFormat,
		repositoryAddress.String(),
		imageName,
	)

	image.SetTag(imageRef)

	image.SetBuildArg(serverPortKey,
		fmt.Sprint(serverPortValue),
	)

	image.SetBuildArg(statusCodeKey,
		fmt.Sprint(statusCodeValue),
	)

	e = image.Build(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	e = image.Push(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	e = image.Remove()
	if e != nil {
		t.Error(e)
	}

	// start container using image pulled from repository

	config, e = configs.NewDockerContainerConfig(imageRef)
	if e != nil {
		t.Error(e)
	}

	config.PublishTCPPort(
		fmt.Sprint(serverPortValue),
		localhost,
		fmt.Sprint(serverPortValue),
	)

	container, e = containers.NewDockerContainer(
		imageName,
		config,
		os.Stderr,
	)
	if e != nil {
		t.Error(e)
	}

	defer container.Remove()

	// interact with server in container via client

	client, e = clients.NewHTTPClient()
	if e != nil {
		t.Error(e)
	}

	serverAddress.Port = serverPortValue

	endpoint = url.URL{
		Scheme: scheme,
		Host:   serverAddress.String(),
	}

	status, e = client.GetStatusCodeFromEndpoint(endpoint, timeout)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, statusCodeValue, status)
}
