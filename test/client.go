package test

import (
	"github.com/joel-ling/alduin/test/clients"
)

type Client interface {
	SendRequestToServerEndpoint() (int, error)
}

func NewClient() (Client, error) {
	return clients.NewHTTPStatusCodeGetter()
}
