package image

import (
	"testing"

	"github.com/joel-ling/lookout/docker"
)

func TestImage(t *testing.T) {
	const (
		emptyString       = ""
		expectedImageHash = "sha256:" +
			"27e20913aa2ea03fed17242709bed7468771c025752c5b9d6e5a8ab8553ac6a7"
		imageNameTag = "bitnami/kubectl:1.21.4"
	)

	var (
		e         error
		image     *Image
		retriever ImageHashRetriever
		updated   bool
	)

	image = NewImage(imageNameTag)

	retriever, _, e = docker.NewClient(false,
		emptyString,
		emptyString,
		emptyString,
	)
	if e != nil {
		t.Error(e)
	}

	if image.Hash() != emptyString {
		t.Log()
		t.Fail()
	}

	updated, e = image.CheckForUpdate(retriever)
	if e != nil {
		t.Error(e)
	}

	if !updated {
		t.Log()
		t.Fail()
	}

	if image.Hash() != expectedImageHash {
		t.Log()
		t.Fail()
	}

	updated, e = image.CheckForUpdate(retriever)
	if e != nil {
		t.Error(e)
	}

	if updated {
		t.Log()
		t.Fail()
	}
}
