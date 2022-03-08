package repositories

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestRepositoryCredentialsGetter(t *testing.T) {
	const (
		jsonFormat1Credential = `` +
			`{"auths":{"%s":{"username":"%s","password":"%s"}}}`
		jsonFormat2Credentials = `` +
			`{"auths":{` +
			`"%s":{"username":"%s","password":"%s"},` +
			`"%s":{"username":"%s","password":"%s"}` +
			`}}`

		nCredentials = 4

		addressFormat  = "127.0.0.%d:5000"
		passwordFormat = "password%d"
		usernameFormat = "username%d"
	)

	var (
		addresses []string
		passwords []string
		usernames []string

		secrets []v1.Secret

		getter RepositoryCredentialsGetter

		password string
		username string

		e error
		i int
	)

	addresses = make([]string, nCredentials)
	usernames = make([]string, nCredentials)
	passwords = make([]string, nCredentials)

	for i = 0; i < nCredentials; i++ {
		addresses[i] = fmt.Sprintf(addressFormat, i+1)
		usernames[i] = fmt.Sprintf(usernameFormat, i+1)
		passwords[i] = fmt.Sprintf(passwordFormat, i+1)
	}

	secrets = []v1.Secret{
		{
			Data: map[string][]byte{
				v1.DockerConfigJsonKey: []byte(
					fmt.Sprintf(jsonFormat1Credential,
						addresses[0],
						usernames[0],
						passwords[0],
					),
				),
			},
			//Type:
		},
		{
			Data: map[string][]byte{
				v1.DockerConfigJsonKey: []byte(
					fmt.Sprintf(jsonFormat2Credentials,
						addresses[1],
						usernames[1],
						passwords[1],

						addresses[2],
						usernames[2],
						passwords[2],
					),
				),
			},
			Type: v1.SecretTypeDockerConfigJson,
		},
		{
			Data: map[string][]byte{
				v1.DockerConfigJsonKey: []byte(
					fmt.Sprintf(jsonFormat1Credential,
						addresses[3],
						usernames[3],
						passwords[3],
					),
				),
			},
			Type: v1.SecretTypeDockerConfigJson,
		},
	}

	getter, e = NewRepositoryCredentialsGetterFromKubernetesSecrets(secrets)
	if e != nil {
		t.Error(e)
	}

	username, password = getter.GetRepositoryCredentials(addresses[0])

	assert.Equal(t, "", username)
	assert.Equal(t, "", password)

	for i = 1; i < nCredentials; i++ {
		username, password = getter.GetRepositoryCredentials(addresses[i])

		assert.Equal(t, usernames[i], username)
		assert.Equal(t, passwords[i], password)
	}
}
