package docker

import (
	"context"
	"fmt"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/credentials"
	configtypes "github.com/docker/cli/cli/config/types"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/cli/cli/manifest/types"
	"github.com/docker/cli/cli/registry/client"
	"github.com/docker/distribution/reference"
	apitypes "github.com/docker/docker/api/types"
	apitypesregistry "github.com/docker/docker/api/types/registry"
)

type Client struct {
	registryClient client.RegistryClient
}

func NewClient(ctx context.Context, address, username, password string) (
	docker *Client, e error,
) {
	const (
		addressFormat  = "https://%s"
		allowInsecure  = false
		emptyString    = ""
		loginSucceeded = "Login Succeeded"
	)

	var (
		authConfig       apitypes.AuthConfig
		clientOptions    *flags.ClientOptions = flags.NewClientOptions()
		credentialsStore credentials.Store
		dockerCLI        *command.DockerCli
		response         apitypesregistry.AuthenticateOKBody
	)

	dockerCLI, e = command.NewDockerCli()
	if e != nil {
		return
	}

	e = dockerCLI.Initialize(clientOptions)
	if e != nil {
		return
	}

	authConfig = apitypes.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: fmt.Sprintf(addressFormat, address),
	}

	response, e = dockerCLI.Client().RegistryLogin(ctx, authConfig)
	if e != nil {
		return
	}

	if response.Status != loginSucceeded {
		e = fmt.Errorf(response.Status)

		return
	}

	credentialsStore = dockerCLI.ConfigFile().GetCredentialsStore(address)

	e = credentialsStore.Store(
		configtypes.AuthConfig(authConfig),
	)
	if e != nil {
		return
	}

	docker = &Client{
		registryClient: dockerCLI.RegistryClient(allowInsecure),
	}

	return
}

func (docker *Client) RetrieveImageHash(
	ctx context.Context, imageNameTag string,
) (
	imageHash string, e error,
) {
	var (
		imageManifest  types.ImageManifest
		namedReference reference.Named
	)

	namedReference, e = reference.ParseNormalizedNamed(imageNameTag)
	if e != nil {
		return
	}

	imageManifest, e = docker.registryClient.GetManifest(ctx, namedReference)
	if e != nil {
		return
	}

	imageHash = string(imageManifest.Descriptor.Digest)

	return
}
