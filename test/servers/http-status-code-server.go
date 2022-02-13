package servers

import (
	"net"
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
	const (
		network = "tcp"
	)

	var (
		listener net.Listener
		timer    *time.Timer
	)

	s = &httpStatusCodeServer{
		httpServer: http.Server{
			Handler: http.HandlerFunc(
				func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(statusCode)
				},
			),
		},
	}

	listener, e = net.Listen(network, constants.StatusCodeServerListenAddress)
	if e != nil {
		return
	}

	go s.httpServer.Serve(listener)

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
