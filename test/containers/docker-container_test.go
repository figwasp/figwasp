package containers

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joel-ling/alduin/pkg/clients"
	"github.com/joel-ling/alduin/pkg/containers"
	"github.com/joel-ling/alduin/pkg/containers/configs"
)

func TestDockerContainer(t *testing.T) {
	const (
		containerPort = "80"
		hostIP        = "127.0.0.1"
		hostPort      = 8001
		imageRef      = "nginx"
		scheme        = "http"
		timeout       = time.Second
	)

	var (
		address   net.TCPAddr
		client    *clients.HTTPClient
		config    *configs.DockerContainerConfig
		container *containers.DockerContainer
		endpoint  url.URL

		status int

		e error
	)

	config, e = configs.NewDockerContainerConfig(imageRef)
	if e != nil {
		t.Error(e)
	}

	address = net.TCPAddr{
		IP:   net.ParseIP(hostIP),
		Port: hostPort,
	}

	config.PublishTCPPort(containerPort,
		address.IP.String(),
		fmt.Sprint(address.Port),
	)

	container, e = containers.NewDockerContainer(imageRef, imageRef,
		config,
		os.Stderr,
	)
	if e != nil {
		t.Error(e)
	}

	defer container.Remove()

	client, e = clients.NewHTTPClient()
	if e != nil {
		t.Error(e)
	}

	endpoint = url.URL{
		Scheme: scheme,
		Host:   address.String(),
	}

	status, e = client.GetStatusCodeFromEndpoint(endpoint, timeout)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, http.StatusOK, status)
}
