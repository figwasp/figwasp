package repositories

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"strings"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory"
	"github.com/joel-ling/alduin/test/constants"
	"github.com/joel-ling/alduin/test/images"
)

var (
	//go:embed config.yml
	configTemplate string
)

type dockerRegistry struct {
	server *registry.Registry
	cancel context.CancelFunc
}

func NewDockerRegistry() (r *dockerRegistry, e error) {
	const (
		network = "tcp"
	)

	var (
		config       *configuration.Configuration
		configReader *strings.Reader
		configString string

		contextWithCancel context.Context
	)

	configString = fmt.Sprintf(configTemplate, constants.DockerRegistryAddress)

	configReader = strings.NewReader(configString)

	config, e = configuration.Parse(configReader)
	if e != nil {
		return
	}

	r = &dockerRegistry{}

	contextWithCancel, r.cancel = context.WithCancel(
		context.Background(),
	)

	r.server, e = registry.NewRegistry(
		contextWithCancel,
		config,
	)
	if e != nil {
		return
	}

	go r.server.ListenAndServe()

	for {
		_, e = net.Dial(network, constants.DockerRegistryAddress)
		if e == nil {
			break
		}
	}

	return
}

func (r *dockerRegistry) Destroy() (e error) {
	r.cancel()

	return
}

func (r *dockerRegistry) BuildAndPushServerImage(
	statusCode int, buildContextPath string,
) (
	e error,
) {
	var (
		image interface {
			Build() error
			Push() error
		}
	)

	image, e = images.NewHTTPStatusCodeServerImage(statusCode, buildContextPath)
	if e != nil {
		return
	}

	e = image.Build()
	if e != nil {
		return
	}

	e = image.Push()
	if e != nil {
		return
	}

	return
}

func (r *dockerRegistry) BuildAndPushAlduinImage() (e error) {
	return
}
