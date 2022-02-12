package test

import (
	"github.com/joel-ling/alduin/test/repositories"
)

type Repository interface {
	destroyable

	BuildAndPushServerImage(int) error
	BuildAndPushAlduinImage() error
}

func NewRepository() (Repository, error) {
	return repositories.NewDockerRegistry()
}
