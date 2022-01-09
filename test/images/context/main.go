package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	message string
	path    string
	port    string
)

func main() {
	http.HandleFunc(path, handleRequest)

	log.Fatal(
		http.ListenAndServe(port, nil),
	)
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	// Respond to a HTTP request with a message defined at build time.

	fmt.Fprintf(writer, message)

	return
}
