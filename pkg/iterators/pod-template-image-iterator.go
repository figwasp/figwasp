package iterators

import (
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

type PodTemplateImageIterator interface {
	IterateOverImages(func(string) error) error
}

type podTemplateImageIterator struct {
	imageRefs []string
}

func NewPodTemplateImageIterator(object podTemplateObject) (
	i *podTemplateImageIterator, e error,
) {
	var (
		container coreV1.Container
	)

	i = &podTemplateImageIterator{
		imageRefs: make([]string, 0),
	}

	for _, container = range object.PodTemplateSpec().Spec.Containers {
		i.imageRefs = append(i.imageRefs,
			container.Image,
		)
	}

	return
}

func (i *podTemplateImageIterator) IterateOverImages(
	function func(string) error,
) (
	e error,
) {
	var (
		imageRef string
	)

	for _, imageRef = range i.imageRefs {
		e = function(imageRef)
		if e != nil {
			return
		}
	}

	return
}

type podTemplateObject interface {
	PodTemplateSpec() coreV1.PodTemplateSpec
}

type DeploymentObject appsV1.Deployment

func (o *DeploymentObject) PodTemplateSpec() coreV1.PodTemplateSpec {
	return o.Spec.Template
}
