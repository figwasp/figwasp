package repositories

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	_ "embed"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/inmemory"
)

var (
	//go:embed config.yml
	configTemplateVanilla string

	//go:embed config-auth.yml
	configTemplateBasicAuth string

	//go:embed config-auth-tls.yml
	configTemplateBasicAuthTLS string

	//go:embed htpasswd
	htpasswdBytes []byte
)

type DockerRegistry struct {
	server   *registry.Registry
	cancel   context.CancelFunc
	htpasswd *os.File
}

func NewDockerRegistry(address net.TCPAddr) (*DockerRegistry, error) {
	return newDockerRegistry(address, configTemplateVanilla)
}

func NewDockerRegistryWithBasicAuth(address net.TCPAddr) (
	r *DockerRegistry, e error,
) {
	return newDockerRegistryWithBasicAuth(address, configTemplateBasicAuth)
}

func NewDockerRegistryWithBasicAuthAndTLS(
	address net.TCPAddr, pathToCertificatePEM, pathToPrivateKeyPEM string,
) (
	r *DockerRegistry, e error,
) {
	const (
		placeholder = "%s"
	)

	var (
		configTemplate string
	)

	configTemplate = fmt.Sprintf(configTemplateBasicAuthTLS,
		placeholder,
		pathToCertificatePEM,
		pathToPrivateKeyPEM,
		placeholder,
	)

	r, e = newDockerRegistryWithBasicAuth(address, configTemplate)
	if e != nil {
		return
	}

	return
}

func newDockerRegistryWithBasicAuth(
	address net.TCPAddr, configTemplateBasicAuth string,
) (
	r *DockerRegistry, e error,
) {
	const (
		htpasswdFileDirectory = ""
		htpasswdFilePattern   = "htpasswd-*"
		placeholder           = "%s"
	)

	var (
		configTemplate string
		htpasswdFile   *os.File
	)

	htpasswdFile, e = ioutil.TempFile(
		htpasswdFileDirectory,
		htpasswdFilePattern,
	)
	if e != nil {
		return
	}

	_, e = htpasswdFile.Write(htpasswdBytes)
	if e != nil {
		return
	}

	e = htpasswdFile.Close()
	if e != nil {
		return
	}

	configTemplate = fmt.Sprintf(configTemplateBasicAuth,
		placeholder,
		htpasswdFile.Name(),
	)

	r, e = newDockerRegistry(address, configTemplate)
	if e != nil {
		return
	}

	r.htpasswd = htpasswdFile

	return
}

func newDockerRegistry(address net.TCPAddr, configTemplate string,
) (
	r *DockerRegistry, e error,
) {
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

	e = os.Remove(
		r.htpasswd.Name(),
	)
	if e != nil {
		return
	}

	return
}
