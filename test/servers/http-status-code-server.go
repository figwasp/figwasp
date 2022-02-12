package servers

import (
	"net/http"
	"time"

	"github.com/joel-ling/alduin/test/constants"
)

type httpStatusCodeServer struct {
	httpServer http.Server
}

func NewHTTPStatusCodeServer(statusCode int) (
	s *httpStatusCodeServer, e error,
) {
	var (
		timer *time.Timer
	)

	s = &httpStatusCodeServer{
		httpServer: http.Server{
			Addr: constants.StatusCodeServerAddress,
			Handler: http.HandlerFunc(
				func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(statusCode)
				},
			),
		},
	}

	go s.httpServer.ListenAndServe()

	timer = time.NewTimer(constants.StatusCodeGetterTimeoutDuration)

	for {
		select {
		case <-timer.C:
			return

		default:
			_, e = http.Get(constants.StatusCodeServerEndpointURL)
			if e == nil {
				return
			}
		}
	}
}

func (s *httpStatusCodeServer) Close() (e error) {
	return s.httpServer.Close()
}
