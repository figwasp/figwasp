package repositories

import (
	"context"
	"fmt"
	"net"
	"strings"

	_ "embed"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory"
)

var (
	//go:embed config.yml
	configTemplate string
)

type DockerRegistry struct {
	server *registry.Registry
	cancel context.CancelFunc
}

func NewDockerRegistry(address net.TCPAddr) (r *DockerRegistry, e error) {
	var (
		config       *configuration.Configuration
		configReader *strings.Reader
		configString string

		contextWithCancel context.Context
	)

	configString = fmt.Sprintf(configTemplate,
		address.String(),
	)

	configReader = strings.NewReader(configString)

	config, e = configuration.Parse(configReader)
	if e != nil {
		return
	}

	r = &DockerRegistry{}

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
		_, e = net.Dial(
			address.Network(),
			address.String(),
		)
		if e == nil {
			break
		}
	}

	return
}

func (r *DockerRegistry) Destroy() (e error) {
	r.cancel()

	return
}
