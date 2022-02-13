package constants

import (
	"time"
)

const (
	StatusCodeServerIP   = "127.0.0.1"
	StatusCodeServerPort = "8000"
)

const (
	StatusCodeServerAddress = StatusCodeServerIP + ":" + StatusCodeServerPort
)

const (
	StatusCodeGetterTimeoutDuration = time.Second
	StatusCodeServerContainerName   = "status-code-server"
	StatusCodeServerEndpointURL     = "http://" + StatusCodeServerAddress
)
