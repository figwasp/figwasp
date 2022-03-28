package figwasp

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/juju/errors"
	"github.com/opencontainers/go-digest"
)

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
		pathsToRemove: []string{},
	}

	for _, option = range options {
		e = option(r)
		if e != nil {
			e = errors.Trace(e)

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
		ImageReference types.ImageReference
	)

	ImageReference, e = docker.ParseReference(
		fmt.Sprintf(imageReferenceFormat, imageReferenceString),
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	imageCloser, e = ImageReference.NewImage(ctx, r.systemContext)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	imageManifest, _, e = imageCloser.Manifest(ctx)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	imageDigest, e = manifest.Digest(imageManifest)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	imageDigestString = imageDigest.String()

	return
}

func (r *imageDigestRetriever) Destroy() (e error) {
	var (
		path string
	)

	for _, path = range r.pathsToRemove {
		e = os.RemoveAll(path)
		if e != nil {
			e = errors.Trace(e)

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
			e = errors.Trace(e)

			return
		}

		pathToCACertLink = fmt.Sprintf(pathToCACertLinkFormat, pathToCACertDir)

		e = os.Link(pathToCACert, pathToCACertLink)
		if e != nil {
			e = errors.Trace(e)

			return
		}

		r.systemContext.DockerCertPath = pathToCACertDir

		r.pathsToRemove = append(r.pathsToRemove, pathToCACertDir)

		return
	}

	return
}
