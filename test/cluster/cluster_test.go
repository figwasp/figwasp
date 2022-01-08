package cluster

import (
	"testing"
)

func TestCluster(t *testing.T) {
	const (
		clusterName    = "test-cluster"
		kubeConfigPath = ""
	)

	var (
		cluster Cluster
		e       error
	)

	cluster, e = NewKindCluster(clusterName, kubeConfigPath)
	if e != nil {
		t.Error(e)
	}

	e = cluster.Destroy()
	if e != nil {
		t.Error(e)
	}
}
