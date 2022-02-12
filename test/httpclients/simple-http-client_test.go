package httpclients

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleHTTPClient(t *testing.T) {
	const (
		address   = "127.0.0.1:8000"
		message   = "Hello, World!"
		network   = "tcp"
		timeout   = time.Second
		urlFormat = "http://%s"
	)

	var (
		client   simpleHTTPClient
		ok       chan bool
		server   http.Server
		listener net.Listener

		e error
	)

	client = NewSimpleHTTPClient(
		fmt.Sprintf(urlFormat, address),
		message,
		timeout,
	)

	ok = make(chan bool)

	go client.OK(ok)

	server.Handler = http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			fmt.Fprintf(writer, message)
		},
	)

	listener, e = net.Listen(network, address)
	if e != nil {
		t.Error(e)
	}

	go server.Serve(listener)

	defer server.Close()

	assert.True(t, <-ok)
}
