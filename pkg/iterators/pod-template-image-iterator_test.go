package iterators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

func TestPodTemplateImageIterator(t *testing.T) {
	const (
		imageRef0 = "abc"
		imageRef1 = "def"
		imageRef2 = "ghi"
	)

	var (
		deployment       appsV1.Deployment
		deploymentObject DeploymentObject

		iterator PodTemplateImageIterator

		function          func(string) error
		imageRefs         []string
		imageRefsExpected []string

		e error
	)

	deployment = appsV1.Deployment{
		Spec: appsV1.DeploymentSpec{
			Template: coreV1.PodTemplateSpec{
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{
						{Image: imageRef0},
						{Image: imageRef1},
						{Image: imageRef2},
					},
				},
			},
		},
	}

	deploymentObject = DeploymentObject(deployment)

	iterator, e = NewPodTemplateImageIterator(&deploymentObject)
	if e != nil {
		t.Error(e)
	}

	function = func(imageRef string) (e error) {
		imageRefs = append(imageRefs,
			imageRef,
		)

		return
	}

	e = iterator.IterateOverImages(function)
	if e != nil {
		t.Error(e)
	}

	imageRefsExpected = []string{
		imageRef0,
		imageRef1,
		imageRef2,
	}

	assert.Equal(t,
		imageRefsExpected,
		imageRefs,
	)
}
