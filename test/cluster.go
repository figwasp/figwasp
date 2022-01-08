package test

import (
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"
)

type Cluster interface {
	Destroy() error
}

type kindCluster struct {
	provider       *cluster.Provider
	name           string
	kubeConfigPath string
}

func NewKindCluster(name, kubeConfigPath string) (c *kindCluster, e error) {
	// Initialise a local Kubernetes cluster using Docker containers as nodes,
	// overriding any existing cluster with the same name.

	var (
		logger log.Logger
		option cluster.ProviderOption
	)

	logger = cmd.NewLogger()

	option = cluster.ProviderWithLogger(logger)

	c = &kindCluster{
		provider:       cluster.NewProvider(option),
		name:           name,
		kubeConfigPath: kubeConfigPath,
	}

	e = c.Destroy()
	if e != nil {
		return
	}

	e = c.provider.Create(name)
	if e != nil {
		return
	}

	return
}

func (c *kindCluster) Destroy() (e error) {
	//

	e = c.provider.Delete(
		c.name,
		c.kubeConfigPath,
	)
	if e != nil {
		return
	}

	return
}
