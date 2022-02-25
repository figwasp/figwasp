package images

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"

	"github.com/joel-ling/alduin/test/pkg/credentials"
	"github.com/joel-ling/alduin/test/pkg/images"
	"github.com/joel-ling/alduin/test/pkg/repositories"
)

const (
	imageDigestEncodedLength = 64
)

func TestImageDigestRetrieverAgainstPublicRepository(t *testing.T) {
	const (
		imageRef = "golang:1.17"
	)

	var (
		retriever ImageDigestRetriever

		imageDigest       digest.Digest
		imageDigestString string

		e error
	)

	retriever, e = NewPublicImageDigestRetriever()
	if e != nil {
		t.Error(e)
	}

	imageDigestString, e = retriever.RetrieveImageDigest(
		imageRef,
		context.Background(),
	)
	if e != nil {
		t.Error(e)
	}

	imageDigest = digest.Digest(imageDigestString)

	assert.Equal(t,
		digest.SHA256,
		imageDigest.Algorithm(),
	)

	assert.Equal(t,
		imageDigestEncodedLength,
		len(imageDigest.Encoded()),
	)
}

func TestImageDigestRetrieverAgainstPrivateRepositoryWithBasicAuthAndTLS(
	t *testing.T,
) {
	const (
		repositoryHost = "127.0.0.1"
		repositoryPort = 5000

		buildContextPath = "../.."
		dockerfilePath   = "test/build/dummy/Dockerfile"

		imageName      = "dummy"
		imageRefFormat = "%s:%d/%s"

		username = "username"
		password = "password"
	)

	var (
		repository        *repositories.DockerRegistry
		repositoryAddress net.TCPAddr
		repositoryCert    *credentials.TLSCertificate

		image    *images.DockerImage
		imageRef string

		retriever ImageDigestRetriever

		imageDigest       digest.Digest
		imageDigestString string

		e error
	)

	repositoryAddress = net.TCPAddr{
		IP:   net.ParseIP(repositoryHost),
		Port: repositoryPort,
	}

	repositoryCert, e = credentials.NewTLSCertificateForIPAddress(
		repositoryAddress.IP,
	)
	if e != nil {
		t.Error(e)
	}

	defer repositoryCert.Remove()

	repository, e = repositories.NewDockerRegistryWithBasicAuthAndTLS(
		repositoryAddress,
		repositoryCert.PathToCertificatePEM(),
		repositoryCert.PathToPrivateKeyPEM(),
	)
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath)
	if e != nil {
		t.Error(e)
	}

	imageRef = fmt.Sprintf(imageRefFormat,
		repositoryHost,
		repositoryPort,
		imageName,
	)

	image.SetTag(imageRef)

	e = image.Build(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	e = image.PushWithBasicAuth(
		os.Stderr,
		username,
		password,
	)
	if e != nil {
		t.Error(e)
	}

	retriever, e = NewPrivateImageDigestRetriever(
		username,
		password,
		repositoryCert.PathToCertificatePEM(),
	)
	if e != nil {
		t.Error(e)
	}

	defer retriever.Close()

	imageDigestString, e = retriever.RetrieveImageDigest(
		imageRef,
		context.Background(),
	)
	if e != nil {
		t.Error(e)
	}

	imageDigest = digest.Digest(imageDigestString)

	assert.Equal(t,
		digest.SHA256,
		imageDigest.Algorithm(),
	)

	assert.Equal(t,
		imageDigestEncodedLength,
		len(imageDigest.Encoded()),
	)
}
