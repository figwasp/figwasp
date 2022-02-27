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
	pathsToRemove []string
}

func NewImageDigestRetriever(options ...imageDigestRetrieverOption) (
	r *imageDigestRetriever, e error,
) {
	var (
		option imageDigestRetrieverOption
	)

	r = &imageDigestRetriever{
		systemContext: &types.SystemContext{},
		pathsToRemove: make([]string, 0),
	}

	for _, option = range options {
		e = option(r)
		if e != nil {
			return
		}
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
	var (
		path string
	)

	for _, path = range r.pathsToRemove {
		e = os.RemoveAll(path)
		if e != nil {
			return
		}
	}

	return
}

type imageDigestRetrieverOption func(*imageDigestRetriever) error

func WithBasicAuthentication(username, password string) (
	option imageDigestRetrieverOption,
) {
	option = func(r *imageDigestRetriever) (e error) {
		r.systemContext.DockerAuthConfig = &types.DockerAuthConfig{
			Username: username,
			Password: password,
		}

		return
	}

	return
}

func WithSelfSignedTLSCertificate(pathToCACert string) (
	option imageDigestRetrieverOption,
) {
	option = func(r *imageDigestRetriever) (e error) {
		const (
			pathToCACertDirParent  = ""
			pathToCACertDirPattern = "*"

			pathToCACertLinkFormat = "%s/ca.crt"
			// https://pkg.go.dev/
			//  github.com/containers/image/v5/types#SystemContext
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

		r.systemContext.DockerCertPath = pathToCACertDir

		r.pathsToRemove = append(r.pathsToRemove, pathToCACertDir)

		return
	}

	return
}
