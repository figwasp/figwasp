package images

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"

	"github.com/figwasp/figwasp/test/pkg/credentials"
	"github.com/figwasp/figwasp/test/pkg/images"
	"github.com/figwasp/figwasp/test/pkg/repositories"
)

func TestImageDigestRetrieverAgainstPublicRepository(t *testing.T) {
	const (
		imageRef0 = "golang:1.5.1"
		// last updated 2015; Schema 1 manifest outdated
		imageRef1 = "golang:1.10.1"
		// last updated 2018; Schema 2 manifest

		// obtained by pulling and inspecting images
		imageDigestEncodedExpected0 = "" +
			"23ca2c13e498feab91c5fa38f56d3b2bfaebc98d41d283102b467a2900e48e40"
		imageDigestEncodedExpected1 = "" +
			"4826b5c314a498142c7291ad835ab6be1bf02f7813d6932d01f1f0f1383cdda1"
	)

	var (
		testCases map[string]string

		retriever *imageDigestRetriever

		imageDigestEncodedExpected string
		imageRef                   string

		imageDigest       digest.Digest
		imageDigestString string

		e error
	)

	testCases = map[string]string{
		imageRef0: imageDigestEncodedExpected0,
		imageRef1: imageDigestEncodedExpected1,
	}

	retriever, e = NewImageDigestRetriever()
	if e != nil {
		t.Error(e)
	}

	for imageRef, imageDigestEncodedExpected = range testCases {
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
			imageDigestEncodedExpected,
			imageDigest.Encoded(),
		)
	}
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

		imageDigestEncodedLength = 64
	)

	var (
		credential *credentials.TLSCertificate

		repository        *repositories.DockerRegistry
		repositoryAddress net.TCPAddr

		image    *images.DockerImage
		imageRef string

		retriever *imageDigestRetriever

		imageDigest       digest.Digest
		imageDigestString string

		e error
	)

	credential, e = credentials.NewTLSCertificate(
		credentials.WithExtendedKeyUsageForServerAuth(),
		credentials.WithIPAddress(repositoryHost),
	)
	if e != nil {
		t.Error(e)
	}

	defer credential.Destroy()

	repositoryAddress = net.TCPAddr{
		IP:   net.ParseIP(repositoryHost),
		Port: repositoryPort,
	}

	repository, e = repositories.NewDockerRegistry(repositoryAddress,
		repositories.WithBasicAuthentication(username, password),
		repositories.WithTransportLayerSecurity(
			credential.PathToCertPEM(),
			credential.PathToKeyPEM(),
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	imageRef = fmt.Sprintf(imageRefFormat,
		repositoryHost,
		repositoryPort,
		imageName,
	)

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath,
		images.WithTag(imageRef),
		images.WithOutputStream(os.Stderr),
	)
	if e != nil {
		t.Error(e)
	}

	defer image.Destroy()

	e = image.PushWithBasicAuth(os.Stderr, username, password)
	if e != nil {
		t.Error(e)
	}

	retriever, e = NewImageDigestRetriever(
		WithBasicAuthentication(username, password),
		WithSelfSignedTLSCertificate(
			credential.PathToCertPEM(),
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer retriever.Destroy()

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
