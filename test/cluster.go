package test

import (
	"github.com/joel-ling/alduin/test/clusters"
)

type Cluster interface {
	destroyable

	DeployServer() error
	DeployAlduin() error
}

func NewCluster() (Cluster, error) {
	return clusters.NewKindCluster()
}
