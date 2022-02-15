package containers

import (
	"net/http"
	"testing"

	"github.com/joel-ling/alduin/test"
	"github.com/joel-ling/alduin/test/clients"
	"github.com/stretchr/testify/assert"
)

func TestDockerContainer(t *testing.T) {
	const (
		containerPort = "80"
		imageRef      = "nginx:latest"
	)

	var (
		client    test.Client
		container *dockerContainer

		status int

		e error
	)

	container, e = NewDockerContainer(imageRef, containerPort)
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

	assert.EqualValues(t, http.StatusOK, status)
}
