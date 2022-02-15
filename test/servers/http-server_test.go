package servers

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joel-ling/alduin/pkg/clients"
	"github.com/joel-ling/alduin/pkg/servers"
)

func TestHTTPServer(t *testing.T) {
	const (
		status0 = http.StatusTeapot
		timeout = time.Second
	)

	var (
		client *clients.HTTPClient
		server *servers.HTTPServer

		address net.TCPAddr
		status  int

		e error
	)

	server, e = servers.NewHTTPServer()
	if e != nil {
		t.Error(e)
	}

	e = server.ServeStatusCodeAtAddress(address, status0)
	if e != nil {
		t.Error(e)
	}

	client, e = clients.NewHTTPClient()
	if e != nil {
		t.Error(e)
	}

	status, e = client.GetStatusCodeFromEndpoint(
		server.Endpoint(),
		timeout,
	)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, status0, status)
}
