package clients

import (
	"io"
	"net/http"
	"testing"

	"github.com/joel-ling/alduin/test/servers"
	"github.com/stretchr/testify/assert"
)

func TestHTTPStatusCodeGetter(t *testing.T) {
	const (
		status0 = http.StatusTeapot
	)

	var (
		client *httpStatusCodeGetter
		server io.Closer

		status int

		e error
	)

	client, e = NewHTTPStatusCodeGetter()
	if e != nil {
		t.Error(e)
	}

	server, e = servers.NewHTTPStatusCodeServer(status0)
	if e != nil {
		t.Error(e)
	}

	defer server.Close()

	status, e = client.SendRequestToServerEndpoint()
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, status0, status)
}
