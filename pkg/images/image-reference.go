package images

import (
	"github.com/containers/image/v5/docker/reference"
)

type ImageReference interface {
	RepositoryAddress() string
	NamedAndTagged() string
	ImageDigest() string
}

type imageReference struct {
	repositoryAddress string
	namedAndTagged    string
	imageDigest       string
}

func NewImageReferenceFromCanonicalString(s string) (
	r *imageReference, e error,
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
		return
	}

	r = &imageReference{
		repositoryAddress: reference.Domain(named),
		namedAndTagged:    namedTagged.String(),
		imageDigest:       named.(reference.Digested).Digest().String(),
	}

	return
}

func (r *imageReference) RepositoryAddress() string {
	return r.repositoryAddress
}

func (r *imageReference) NamedAndTagged() string {
	return r.namedAndTagged
}

func (r *imageReference) ImageDigest() string {
	return r.imageDigest
}
