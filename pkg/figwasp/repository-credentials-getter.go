package figwasp

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/cmd/create"
)

type repositoryCredentialsGetter struct {
	store map[string]credentials
}

func NewRepositoryCredentialsGetterFromKubernetesSecrets(secrets []v1.Secret) (
	g *repositoryCredentialsGetter, e error,
) {
	var (
		secret v1.Secret

		dockerConfig      create.DockerConfigJSON
		dockerConfigEntry create.DockerConfigEntry
		repositoryAddress string
	)

	g = &repositoryCredentialsGetter{
		store: make(map[string]credentials),
	}

	for _, secret = range secrets {
		if secret.Type != v1.SecretTypeDockerConfigJson {
			continue
		}

		e = json.Unmarshal(
			secret.Data[v1.DockerConfigJsonKey],
			&dockerConfig,
		)
		if e != nil {
			return
		}

		for repositoryAddress, dockerConfigEntry = range dockerConfig.Auths {
			g.store[repositoryAddress] = credentials{
				username: dockerConfigEntry.Username,
				password: dockerConfigEntry.Password,
			}
		}
	}

	return
}

func (g *repositoryCredentialsGetter) GetRepositoryCredentials(
	repositoryAddress string,
) (
	username, password string,
) {
	var (
		entry credentials
		found bool
	)

	entry, found = g.store[repositoryAddress]
	if !found {
		return
	}

	username, password = entry.username, entry.password

	return
}

type credentials struct {
	username string
	password string
}
