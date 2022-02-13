package main

import (
	"log"
	"strconv"

	"github.com/joel-ling/alduin/test/servers"
)

var (
	statusCodeString string
	// The Go linker supports build-time variable injection of strings only.
	// https://pkg.go.dev/cmd/link
)

func main() {
	var (
		status int

		e error
	)

	status, e = strconv.Atoi(statusCodeString)
	if e != nil {
		log.Fatalln(e)
	}

	_, e = servers.NewHTTPStatusCodeServer(status)
	if e != nil {
		log.Fatalln(e)
	}

	for {
	}
}
