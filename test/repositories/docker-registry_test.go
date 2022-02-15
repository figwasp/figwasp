package repositories

import (
	"net/http"
	"testing"

	"github.com/joel-ling/alduin/test/clients"
	"github.com/joel-ling/alduin/test/constants"
	"github.com/joel-ling/alduin/test/containers"
	"github.com/stretchr/testify/assert"
)

func TestDockerRegistry(t *testing.T) {
	const (
		buildContextPath = "../images/http-status-code-server"
		status0          = http.StatusTeapot
	)

	var (
		client    interface{ GetStatusCodeFromEndpoint() (int, error) }
		container interface{ Destroy() error }
		registry  *dockerRegistry

		status int

		e error
	)

	registry, e = NewDockerRegistry()
	if e != nil {
		t.Error(e)
	}

	defer registry.Destroy()

	e = registry.BuildAndPushServerImage(status0, buildContextPath)
	if e != nil {
		t.Error(e)
	}

	container, e = containers.NewDockerContainer(
		constants.StatusCodeServerImageRef,
		constants.StatusCodeServerContainerPort,
	)
	if e != nil {
		t.Error(e)
	}

	defer container.Destroy()

	client, e = clients.NewHTTPStatusCodeGetter()
	if e != nil {
		t.Error(e)
	}

	status, e = client.GetStatusCodeFromEndpoint()
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, status0, status)
}
