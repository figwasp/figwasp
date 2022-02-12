package servers

import (
	"net/http"
	"testing"

	"github.com/joel-ling/alduin/test/constants"
	"github.com/stretchr/testify/assert"
)

func TestHTTPStatusCodeServer(t *testing.T) {
	const (
		statusCode = http.StatusTeapot
	)

	var (
		response *http.Response
		server   *httpStatusCodeServer

		e error
	)

	server, e = NewHTTPStatusCodeServer(statusCode)
	if e != nil {
		t.Error(e)
	}

	defer server.Close()

	response, e = http.Get(constants.StatusCodeServerEndpointURL)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, statusCode, response.StatusCode)
}
