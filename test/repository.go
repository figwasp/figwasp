package test

type Repository interface {
	destroyable

	BuildAndPushServerImage(int) error
	BuildAndPushAlduinImage() error
}

func NewRepository() (r Repository, e error) {
	return
}
