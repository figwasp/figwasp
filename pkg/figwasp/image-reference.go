package figwasp

import (
	"github.com/containers/image/v5/docker/reference"
	"github.com/juju/errors"
)

type ImageReference struct {
	RepositoryAddress string
	NamedAndTagged    string
	ImageDigest       string
}

func NewImageReferenceFromCanonicalString(s string) (
	r ImageReference, e error,
) {
	const (
		defaultTag = "latest"
	)

	var (
		named       reference.Named
		namedTagged reference.NamedTagged
		ok          bool
		tag         string
		tagged      reference.Tagged
	)

	named, e = reference.ParseNamed(s)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	tagged, ok = named.(reference.Tagged)

	if ok {
		tag = tagged.Tag()

	} else {
		tag = defaultTag
	}

	namedTagged, e = reference.WithTag(
		reference.TrimNamed(named), // remove digest and tag
		tag,                        // recover tag
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	r = ImageReference{
		RepositoryAddress: reference.Domain(named),
		NamedAndTagged:    namedTagged.String(),
		ImageDigest:       named.(reference.Digested).Digest().String(),
	}

	return
}
