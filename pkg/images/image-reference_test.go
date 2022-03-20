package images

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageReference(t *testing.T) {
	const (
		repositoryAddress = "docker.io"
		namedAndTagged    = repositoryAddress + "/library/busybox:latest"
		imageDigest       = "sha256:" +
			"7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa"

		canonicalString = namedAndTagged + "@" + imageDigest
	)

	var (
		reference ImageReference

		e error
	)

	reference, e = NewImageReferenceFromCanonicalString(canonicalString)
	if e != nil {
		t.Error(e)
	}

	assert.Equal(t,
		repositoryAddress,
		reference.RepositoryAddress(),
	)

	assert.Equal(t,
		namedAndTagged,
		reference.NamedAndTagged(),
	)

	assert.Equal(t,
		imageDigest,
		reference.ImageDigest(),
	)
}
