package test

type Cluster interface {
	destroyable

	DeployServer() error
	DeployAlduin() error
}

func NewCluster() (c Cluster, e error) {
	return
}
