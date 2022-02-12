package clients

import (
	"time"
)

type httpStatusCodeGetter struct{}

func NewHTTPStatusCodeGetter() (g *httpStatusCodeGetter, e error) {
	g = &httpStatusCodeGetter{}

	return
}

func (g *httpStatusCodeGetter) SendRequestToServerEndpoint(
	timeout time.Duration,
) (
	statusCode int, e error,
) {
	return
}
