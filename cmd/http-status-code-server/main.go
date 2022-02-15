package main

import (
	"fmt"
	"log"
	"net"

	"github.com/joel-ling/alduin/pkg/servers"
)

var (
	serverPortString string
	statusCodeString string
	// The Go linker supports build-time variable injection of strings only.
	// https://pkg.go.dev/cmd/link
)

func main() {
	const (
		serverPortFormat = "%d"
		statusCodeFormat = "%d"
	)

	var (
		address net.TCPAddr
		server  *servers.HTTPServer

		statusCode int

		e error
	)

	_, e = fmt.Sscanf(statusCodeString, statusCodeFormat, &statusCode)
	if e != nil {
		log.Fatalln(e)
	}

	_, e = fmt.Sscanf(serverPortString, serverPortFormat, &address.Port)
	if e != nil {
		log.Fatalln(e)
	}

	server, e = servers.NewHTTPServer()
	if e != nil {
		log.Fatalln(e)
	}

	e = server.ServeStatusCodeAtAddress(address, statusCode)
	if e != nil {
		log.Fatalln(e)
	}

	defer server.Close()

	for {
	}
}
