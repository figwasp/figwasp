package constants

import (
	"time"
)

const (
	localhost = "127.0.0.1"
)

const (
	StatusCodeServerIP   = localhost
	StatusCodeServerPort = "8000"
)

const (
	StatusCodeServerAddress = StatusCodeServerIP + ":" + StatusCodeServerPort
)

const (
	DockerRegistryAddress           = localhost + ":5000"
	StatusCodeGetterTimeoutDuration = time.Second
	StatusCodeServerContainerName   = "status-code-server"
	StatusCodeServerContainerPort   = StatusCodeServerPort + "/tcp"
	StatusCodeServerEndpointURL     = "http://" + StatusCodeServerAddress
	StatusCodeServerListenAddress   = "0.0.0.0:" + StatusCodeServerPort
)

const (
	StatusCodeServerImageRef = DockerRegistryAddress + "/" +
		StatusCodeServerContainerName + ":latest"
)
