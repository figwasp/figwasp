package clients

import (
	"net/http"
	"time"

	"github.com/joel-ling/alduin/test/constants"
)

type httpStatusCodeGetter struct{}

func NewHTTPStatusCodeGetter() (g *httpStatusCodeGetter, e error) {
	g = &httpStatusCodeGetter{}

	return
}

func (g *httpStatusCodeGetter) GetStatusCodeFromEndpoint() (
	statusCode int, e error,
) {
	var (
		response *http.Response
		timer    *time.Timer
	)

	timer = time.NewTimer(constants.StatusCodeGetterTimeoutDuration)

	for {
		select {
		case <-timer.C:
			return

		default:
			response, e = http.Get(constants.StatusCodeServerEndpointURL)
			if e != nil {
				break
			}

			statusCode = response.StatusCode

			return
		}
	}
}
