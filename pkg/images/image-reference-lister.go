package images

import (
	"k8s.io/api/core/v1"
)

type ImageReferenceLister interface {
	ListImageReferences() []ImageReference
}

type imageReferenceLister struct {
	references map[string]ImageReference
}

func NewImageReferenceListerFromPods(pods []v1.Pod) (
	l *imageReferenceLister, e error,
) {
	var (
		pod v1.Pod

		containerStatus v1.ContainerStatus

		ok bool
	)

	l = &imageReferenceLister{
		references: make(map[string]ImageReference),
	}

	for _, pod = range pods {
		for _, containerStatus = range pod.Status.ContainerStatuses {
			_, ok = l.references[containerStatus.ImageID]
			if ok {
				continue
			}

			l.references[containerStatus.ImageID], e =
				NewImageReferenceFromCanonicalString(containerStatus.ImageID)
			if e != nil {
				return
			}
		}
	}

	return
}

func (l *imageReferenceLister) ListImageReferences() (list []ImageReference) {
	var (
		reference ImageReference

		i int
	)

	list = make([]ImageReference,
		len(l.references),
	)

	for _, reference = range l.references {
		list[i] = reference

		i++
	}

	return
}
