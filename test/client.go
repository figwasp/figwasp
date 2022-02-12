package test

import (
	"time"

	"github.com/joel-ling/alduin/test/clients"
)

type Client interface {
	SendRequestToServerEndpoint(time.Duration) (int, error)
}

func NewClient() (Client, error) {
	return clients.NewHTTPStatusCodeGetter()
}
