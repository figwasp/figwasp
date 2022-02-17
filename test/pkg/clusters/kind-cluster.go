package clusters

import (
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"
)

type KindCluster struct {
	provider       *cluster.Provider
	name           string
	kubeConfigPath string
}

func NewKindCluster(nodeImageRef, name, kubeConfigPath string) (
	c *KindCluster, e error,
) {
	var (
		createOption   cluster.CreateOption
		logger         log.Logger
		providerOption cluster.ProviderOption
	)

	logger = cmd.NewLogger()

	providerOption = cluster.ProviderWithLogger(logger)

	c = &KindCluster{
		provider:       cluster.NewProvider(providerOption),
		name:           name,
		kubeConfigPath: kubeConfigPath,
	}

	e = c.Destroy()
	if e != nil {
		return
	}

	createOption = cluster.CreateWithNodeImage(nodeImageRef)

	e = c.provider.Create(name, createOption)
	if e != nil {
		return
	}

	return
}

func (c *KindCluster) Destroy() (e error) {
	e = c.provider.Delete(
		c.name,
		c.kubeConfigPath,
	)
	if e != nil {
		return
	}

	return
}
