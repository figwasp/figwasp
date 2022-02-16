package clients

import (
	"net/http"
	"net/url"
	"time"
)

type HTTPClient struct{}

func NewHTTPClient() (c *HTTPClient, e error) {
	c = &HTTPClient{}

	return
}

func (c *HTTPClient) GetStatusCodeFromEndpoint(
	locator url.URL, timeout time.Duration,
) (
	statusCode int, e error,
) {
	var (
		response *http.Response
		timer    *time.Timer
	)

	timer = time.NewTimer(timeout)

	for {
		select {
		case <-timer.C:
			return

		default:
			response, e = http.Get(
				locator.String(),
			)
			if e != nil {
				break
			}

			statusCode = response.StatusCode

			return
		}
	}

	return
}
