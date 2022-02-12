package constants

import (
	"time"
)

const (
	StatusCodeServerAddress = "127.0.0.1:8000"
)

const (
	StatusCodeGetterTimeoutDuration = time.Second
	StatusCodeServerEndpointURL     = "http://" + StatusCodeServerAddress
)
