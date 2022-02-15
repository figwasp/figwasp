package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joel-ling/alduin/pkg/clients"
)

func TestMain(t *testing.T) {
	const (
		port    = 8000
		scheme  = "http"
		status0 = http.StatusTeapot
		timeout = time.Second
	)

	var (
		address net.TCPAddr
		client  *clients.HTTPClient

		status int

		e error
	)

	serverPortString = fmt.Sprint(port)
	statusCodeString = fmt.Sprint(status0)

	client, e = clients.NewHTTPClient()
	if e != nil {
		t.Error(e)
	}

	go main()

	address.Port = port

	status, e = client.GetStatusCodeFromEndpoint(
		url.URL{
			Scheme: scheme,
			Host:   address.String(),
		},
		timeout,
	)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, status0, status)
}
