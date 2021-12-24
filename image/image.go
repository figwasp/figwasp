package image

type Image struct {
	nameTag string
	hashOld string
	hashNew string
}

func NewImage(nameTag string) *Image {
	return &Image{nameTag: nameTag}
}

func (image *Image) Hash() string {
	return image.hashNew
}

func (image *Image) retrieveHash(retriever ImageHashRetriever) (e error) {
	image.hashNew, e = retriever.RetrieveImageHash(image.nameTag)
	if e != nil {
		return
	}

	return
}

func (image *Image) CheckForUpdate(retriever ImageHashRetriever) (
	updated bool, e error,
) {
	e = image.retrieveHash(retriever)
	if e != nil {
		return
	}

	if image.hashNew == image.hashOld {
		return

	} else {
		updated = true

		image.hashOld = image.hashNew
	}

	return
}
