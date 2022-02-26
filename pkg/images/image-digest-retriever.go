package images

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
)

type ImageDigestRetriever interface {
	io.Closer

	RetrieveImageDigest(string, context.Context) (string, error)
}

type imageDigestRetriever struct {
	systemContext *types.SystemContext
}

func NewPublicImageDigestRetriever() (
	r *imageDigestRetriever, e error,
) {
	r = &imageDigestRetriever{}

	return
}

func NewPrivateImageDigestRetriever(
	username, password, pathToCACert string,
) (
	r *imageDigestRetriever, e error,
) {
	const (
		pathToCACertDirParent  = ""
		pathToCACertDirPattern = "*"

		pathToCACertLinkFormat = "%s/ca.crt"
		// https://pkg.go.dev/github.com/containers/image/v5/types#SystemContext
		// > a directory containing a CA certificate (ending with ".crt")
	)

	var (
		pathToCACertDir  string
		pathToCACertLink string
	)

	pathToCACertDir, e = ioutil.TempDir(
		pathToCACertDirParent,
		pathToCACertDirPattern,
	)
	if e != nil {
		return
	}

	pathToCACertLink = fmt.Sprintf(pathToCACertLinkFormat, pathToCACertDir)

	e = os.Link(pathToCACert, pathToCACertLink)
	if e != nil {
		return
	}

	r = &imageDigestRetriever{
		systemContext: &types.SystemContext{
			DockerCertPath: pathToCACertDir,
			DockerAuthConfig: &types.DockerAuthConfig{
				Username: username,
				Password: password,
			},
		},
	}

	return
}

func (r *imageDigestRetriever) RetrieveImageDigest(
	imageReferenceString string, ctx context.Context,
) (
	imageDigestString string, e error,
) {
	const (
		imageReferenceFormat = "//%s"
	)

	var (
		imageCloser    types.ImageCloser
		imageDigest    digest.Digest
		imageManifest  []byte
		imageReference types.ImageReference
	)

	imageReference, e = docker.ParseReference(
		fmt.Sprintf(imageReferenceFormat, imageReferenceString),
	)
	if e != nil {
		return
	}

	imageCloser, e = imageReference.NewImage(ctx, r.systemContext)
	if e != nil {
		return
	}

	imageManifest, _, e = imageCloser.Manifest(ctx)
	if e != nil {
		return
	}

	imageDigest, e = manifest.Digest(imageManifest)
	if e != nil {
		return
	}

	imageDigestString = imageDigest.String()

	return
}

func (r *imageDigestRetriever) Close() (e error) {
	if r.systemContext != nil && len(r.systemContext.DockerCertPath) > 0 {
		e = os.RemoveAll(r.systemContext.DockerCertPath)
		if e != nil {
			return
		}
	}

	return
}
