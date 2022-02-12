package httpclients

import (
	"io"
	"net/http"
	"time"
)

type simpleHTTPClient struct {
	url     string
	message string
	timeout time.Duration
}

func NewSimpleHTTPClient(url, message string, timeout time.Duration) (
	c simpleHTTPClient,
) {
	c = simpleHTTPClient{
		url:     url,
		message: message,
		timeout: timeout,
	}

	return
}

func (c simpleHTTPClient) OK(ok chan<- bool) {
	var (
		bodyBytes []byte
		response  *http.Response
		timer     *time.Timer

		e error
	)

	timer = time.NewTimer(c.timeout)

	for {
		select {
		case <-timer.C:
			ok <- false

			return

		default:
			response, e = http.Get(c.url)
			if e != nil {
				break
			}

			if response.StatusCode != http.StatusOK {
				break
			}

			bodyBytes, e = io.ReadAll(response.Body)
			if e != nil {
				break
			}

			response.Body.Close()

			if string(bodyBytes) == c.message {
				ok <- true

				return
			}
		}
	}
}
