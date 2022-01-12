package clusters

import (
	"testing"
)

func TestKindCluster(t *testing.T) {
	const (
		clusterName    = "test-cluster"
		kubeConfigPath = ""
		nodeImage      = "kindest/node:v1.21.1"
	)

	var (
		cluster *kindCluster
		e       error
	)

	cluster, e = NewKindCluster(nodeImage, clusterName, kubeConfigPath)
	if e != nil {
		t.Error(e)
	}

	e = cluster.Destroy()
	if e != nil {
		t.Error(e)
	}
}
