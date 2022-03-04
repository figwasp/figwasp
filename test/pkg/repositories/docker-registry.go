package repositories

import (
	"context"
	"net"

	_ "embed"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory"

	"github.com/joel-ling/alduin/test/pkg/hashes"
)

const (
	auth    = "htpasswd"
	pathKey = "path"
)

type DockerRegistry struct {
	config *configuration.Configuration
	cancel context.CancelFunc
	passwd *hashes.HtpasswdFile
}

func NewDockerRegistry(address net.TCPAddr, options ...dockerRegistryOption) (
	r *DockerRegistry, e error,
) {
	const (
		storage = "inmemory"
		version = "0.1"
	)

	var (
		contextWithCancel context.Context
		option            dockerRegistryOption
		server            *registry.Registry
	)

	r = &DockerRegistry{
		config: &configuration.Configuration{
			Version: version,
			Storage: make(map[string]configuration.Parameters),
			Auth:    make(map[string]configuration.Parameters),
		},
	}

	r.config.HTTP.Addr = address.String()

	r.config.Storage[storage] = configuration.Parameters{}

	contextWithCancel, r.cancel = context.WithCancel(
		context.Background(),
	)

	for _, option = range options {
		e = option(r)
		if e != nil {
			return
		}
	}

	server, e = registry.NewRegistry(contextWithCancel, r.config)
	if e != nil {
		return
	}

	go server.ListenAndServe()

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

	if r.passwd != nil {
		e = r.passwd.Destroy()
		if e != nil {
			return
		}
	}

	return
}

type dockerRegistryOption func(*DockerRegistry) error

func WithBasicAuthentication(username, password string) (
	option dockerRegistryOption,
) {
	const (
		realmKey   = "realm"
		realmValue = "basic"
	)

	option = func(r *DockerRegistry) (e error) {
		r.passwd, e = hashes.NewHtpasswdFile(username, password)
		if e != nil {
			return
		}

		r.config.Auth[auth] = configuration.Parameters{
			realmKey: realmValue,
			pathKey:  r.passwd.Path(),
		}

		return
	}

	return
}

func WithTransportLayerSecurity(pathToCertPEM, pathToKeyPEM string) (
	option dockerRegistryOption,
) {
	option = func(r *DockerRegistry) (e error) {
		r.config.HTTP.TLS.Certificate = pathToCertPEM

		r.config.HTTP.TLS.Key = pathToKeyPEM

		return
	}

	return
}
