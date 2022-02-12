package httpclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleHTTPClient(t *testing.T) {
	const (
		message = "Hello, World!"
	)

	var (
		client   simpleHTTPClient
		handler  http.Handler
		received string
		server   *httptest.Server
		status   int

		e error
	)

	handler = http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintf(writer, message)
		},
	)

	server = httptest.NewServer(handler)

	defer server.Close()

	client = NewSimpleHTTPClient()

	status, received, e = client.Get(server.URL)
	if e != nil {
		t.Error(e)
	}
	if status != http.StatusOK {
		t.Fail()
	}

	assert.Equal(t,
		message, received,
	)
}
