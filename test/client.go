package test

import (
	"time"
)

type Client interface {
	SendRequestToServerEndpoint(time.Duration) (int, error)
}

func NewClient() (c Client, e error) {
	return
}
