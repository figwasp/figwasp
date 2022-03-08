package images

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
)

func TestImageReferenceLister(t *testing.T) {
	const (
		canonicalString0 = "docker.io/library/busybox:latest" +
			"@sha256:" +
			"7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa"
		canonicalString1 = "test:5000/repo:tag" +
			"@sha256:" +
			"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

		repositoryAddress0 = "docker.io"

		nUniqueImages = 2
	)

	var (
		pods []v1.Pod

		lister ImageReferenceLister

		e error
	)

	pods = []v1.Pod{
		{
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						ImageID: canonicalString0,
					},
				},
			},
		},
	}

	lister, e = NewImageReferenceListerFromPods(pods)
	if e != nil {
		t.Error(e)
	}

	assert.Equal(t,
		repositoryAddress0,
		lister.ListImageReferences()[0].RepositoryAddress(),
	)

	pods = append(pods,
		v1.Pod{
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						ImageID: canonicalString0,
					},
					{
						ImageID: canonicalString1,
					},
				},
			},
		},
	)

	lister, e = NewImageReferenceListerFromPods(pods)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t,
		nUniqueImages,
		len(lister.ListImageReferences()),
	)
}
