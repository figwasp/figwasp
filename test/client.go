package test

import (
	"github.com/joel-ling/alduin/test/clients"
)

type Client interface {
	GetStatusCodeFromEndpoint() (int, error)
}

func NewClient() (Client, error) {
	return clients.NewHTTPStatusCodeGetter()
}
