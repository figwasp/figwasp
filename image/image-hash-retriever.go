package image

type ImageHashRetriever interface {
	RetrieveImageHash(imageNameTag string) (imageHash string, e error)
}
