package clusters

import (
	"testing"
)

func TestKindCluster(t *testing.T) {
	const (
		name         = "test-cluster"
		nodeImageRef = "kindest/node:v1.23.3"
	)

	var (
		cluster *KindCluster

		e error
	)

	cluster, e = NewKindCluster(nodeImageRef, name)
	if e != nil {
		return
	}

	e = cluster.Destroy()
	if e != nil {
		return
	}
}
