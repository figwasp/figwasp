package test

import (
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	"github.com/distribution/distribution/v3/configuration"
	"github.com/distribution/distribution/v3/registry"
	_ "github.com/distribution/distribution/v3/registry/auth/htpasswd"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
)

func ServeRegistry(
	address, tlsCertPath, tlsKeyPath, htpasswdPath, storePath string,
) (
	e error,
) {
	// Create, configure and serve an instance of Docker Registry*
	// that implements Transport Layer Security and basic authentication.

	// * See https://docs.docker.com/registry/

	const (
		configTemplate = `
                        version: 0.1
                        http:
                          addr: %s
                          tls:
                            certificate: %s
                            key: %s
                        auth:
                          htpasswd:
                            realm: basic
                            path: %s
                        storage:
                          filesystem:
                            rootdirectory: %s
        `
	)

	var (
		background     context.Context = context.Background()
		config         *configuration.Configuration
		configReader   *strings.Reader
		configString   string
		dockerRegistry *registry.Registry
	)

	configString = fmt.Sprintf(configTemplate,
		address,
		tlsCertPath,
		tlsKeyPath,
		htpasswdPath,
		storePath,
	)

	configReader = strings.NewReader(configString)

	config, e = configuration.Parse(configReader)
	if e != nil {
		return
	}

	dockerRegistry, e = registry.NewRegistry(background, config)
	if e != nil {
		return
	}

	go dockerRegistry.ListenAndServe()

	// Add TLS certificate to system cert pool temporarily#
	// so that it will be trusted by the client.

	// # See https://pkg.go.dev/crypto/x509#SystemCertPool, which states
	//   > Any mutations to the returned pool are not written to disk and
	//   > do not affect any other pool returned by SystemCertPool.

	var (
		systemCertPool *x509.CertPool
		tlsCertBytes   []byte
	)

	tlsCertBytes, e = ioutil.ReadFile(tlsCertPath)
	if e != nil {
		return
	}

	systemCertPool, e = x509.SystemCertPool()
	if e != nil {
		return
	}

	systemCertPool.AppendCertsFromPEM(tlsCertBytes)

	// Wait for Docker Registry to begin listening for connections, then return.

	const (
		protocol = "tcp"
	)

	for {
		_, e = net.Dial(protocol, address)
		if e == nil {
			break
		}
	}

	return
}
