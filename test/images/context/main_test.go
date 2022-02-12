package main

import (
	"testing"
	"time"

	"github.com/joel-ling/alduin/test/httpclients"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	const (
		network = "tcp"

		testMessage = "Hello, World!"
		testPath    = "/"
		testPort    = ":8000"

		testAddress = "127.0.0.1" + testPort
		testURL     = "http://" + testAddress

		testTimeout = time.Second
	)

	type (
		httpResponseVerifier interface {
			OK(chan<- bool)
		}
	)

	var (
		ok       chan bool
		verifier httpResponseVerifier
	)

	message = testMessage
	path = testPath
	port = testPort

	verifier = httpclients.NewSimpleHTTPClient(
		testURL,
		testMessage,
		testTimeout,
	)

	ok = make(chan bool)

	go verifier.OK(ok)

	go main()

	assert.True(t, <-ok)
}
