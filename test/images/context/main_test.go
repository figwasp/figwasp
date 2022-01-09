package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
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

	var (
		e            error
		response     *http.Response
		responseBody []byte
		timer        context.Context
	)

	message = testMessage
	path = testPath
	port = testPort

	go main()

	timer, _ = context.WithTimeout(
		context.Background(),
		testTimeout,
	)

	for {
		_, e = net.Dial(network, testAddress)
		if e == nil {
			break
		}

		if timer.Err() != nil {
			t.Error(e)

			break
		}
	}

	response, e = http.Get(testURL)
	if e != nil {
		t.Error(e)
	}

	if response.StatusCode != http.StatusOK {
		t.Fail()
	}

	defer response.Body.Close()

	responseBody, e = io.ReadAll(response.Body)
	if e != nil {
		t.Error(e)
	}

	if string(responseBody) != testMessage {
		t.Fail()
	}
}
