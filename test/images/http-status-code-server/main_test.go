package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/joel-ling/alduin/test/clients"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	const (
		status0 = http.StatusTeapot
	)

	var (
		client interface{ SendRequestToServerEndpoint() (int, error) }

		status int

		e error
	)

	statusCodeString = fmt.Sprint(status0)

	client, e = clients.NewHTTPStatusCodeGetter()
	if e != nil {
		t.Error(e)
	}

	go main()

	status, e = client.SendRequestToServerEndpoint()
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, status0, status)
}
