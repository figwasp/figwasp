package clusters

import (
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"
)

type kindCluster struct {
	provider       *cluster.Provider
	name           string
	kubeConfigPath string
}

func NewKindCluster(nodeImage, name, kubeConfigPath string) (
	c *kindCluster, e error,
) {
	// Initialise a local Kubernetes cluster using Docker containers as nodes,
	// overriding any existing cluster with the same name.

	var (
		createOption   cluster.CreateOption
		logger         log.Logger
		providerOption cluster.ProviderOption
	)

	logger = cmd.NewLogger()

	providerOption = cluster.ProviderWithLogger(logger)

	c = &kindCluster{
		provider:       cluster.NewProvider(providerOption),
		name:           name,
		kubeConfigPath: kubeConfigPath,
	}

	e = c.Destroy()
	if e != nil {
		return
	}

	createOption = cluster.CreateWithNodeImage(nodeImage)

	e = c.provider.Create(name, createOption)
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
