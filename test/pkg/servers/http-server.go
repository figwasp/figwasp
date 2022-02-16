package servers

import (
	"net"
	"net/http"
	"net/url"
)

type HTTPServer struct {
	endpoint url.URL
	server   http.Server
}

func NewHTTPServer() (s *HTTPServer, e error) {
	s = &HTTPServer{}

	return
}

func (s *HTTPServer) ServeStatusCodeAtAddress(
	address net.TCPAddr, statusCode int,
) (
	e error,
) {
	const (
		scheme = "http"
	)

	var (
		listener net.Listener
	)

	s.server.Handler = http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(statusCode)
		},
	)

	listener, e = net.Listen(
		address.Network(),
		address.String(),
	)
	if e != nil {
		return
	}

	go s.server.Serve(listener)

	s.endpoint = url.URL{
		Scheme: scheme,
		Host:   listener.Addr().String(),
	}

	for {
		_, e = http.Get(
			s.endpoint.String(),
		)
		if e == nil {
			return
		}
	}

	return
}

func (s *HTTPServer) Endpoint() url.URL {
	return s.endpoint
}

func (s *HTTPServer) Close() (e error) {
	return s.server.Close()
}
