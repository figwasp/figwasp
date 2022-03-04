package clusters

import (
	_ "embed"
	"io/ioutil"
	"os"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"
)

var (
	kindClusterProvider *cluster.Provider
)

func init() {
	var (
		logger         log.Logger
		providerOption cluster.ProviderOption
	)

	logger = cmd.NewLogger()

	providerOption = cluster.ProviderWithLogger(logger)

	kindClusterProvider = cluster.NewProvider(providerOption)
}

type KindCluster struct {
	name           string
	kubeconfigPath string
	config         *v1alpha4.Cluster
}

func NewKindCluster(nodeImageRef, name string, options ...kindClusterOption) (
	c *KindCluster, e error,
) {
	const (
		tempFileDirectory = ""
		tempFilePattern   = "*"
	)

	var (
		kubeconfigFile *os.File
		option         kindClusterOption
	)

	kubeconfigFile, e = ioutil.TempFile(tempFileDirectory, tempFilePattern)
	if e != nil {
		return
	}

	c = &KindCluster{
		name:           name,
		kubeconfigPath: kubeconfigFile.Name(),
		config:         new(v1alpha4.Cluster),
	}

	c.config.Nodes = []v1alpha4.Node{
		{
			Role:              v1alpha4.ControlPlaneRole,
			ExtraMounts:       make([]v1alpha4.Mount, 0),
			ExtraPortMappings: make([]v1alpha4.PortMapping, 0),
		},
	}

	for _, option = range options {
		e = option(c)
		if e != nil {
			return
		}
	}

	e = kindClusterProvider.Create(name,
		cluster.CreateWithNodeImage(nodeImageRef),
		cluster.CreateWithKubeconfigPath(c.kubeconfigPath),
		cluster.CreateWithV1Alpha4Config(c.config),
	)
	if e != nil {
		return
	}

	return
}

func (c *KindCluster) KubeconfigPath() string {
	return c.kubeconfigPath
}

func (c *KindCluster) Destroy() (e error) {
	e = kindClusterProvider.Delete(
		c.name,
		c.kubeconfigPath,
	)
	if e != nil {
		return
	}

	e = os.Remove(c.kubeconfigPath)
	if e != nil {
		return
	}

	return
}

type kindClusterOption func(*KindCluster) error

func WithExtraMounts(containerPath, hostPath string) (
	option kindClusterOption,
) {
	const (
		nodeIndex = 0
		readonly  = true
	)

	var (
		mount v1alpha4.Mount
	)

	mount = v1alpha4.Mount{
		ContainerPath: containerPath,
		HostPath:      hostPath,
		Readonly:      readonly,
	}

	option = func(c *KindCluster) (e error) {
		c.config.Nodes[nodeIndex].ExtraMounts = append(
			c.config.Nodes[nodeIndex].ExtraMounts,
			mount,
		)

		return
	}

	return
}

func WithExtraPortMapping(port int32) (option kindClusterOption) {
	const (
		nodeIndex = 0
	)

	var (
		portMapping v1alpha4.PortMapping
	)

	portMapping = v1alpha4.PortMapping{
		ContainerPort: port,
		HostPort:      port,
	}

	option = func(c *KindCluster) (e error) {
		c.config.Nodes[nodeIndex].ExtraPortMappings = append(
			c.config.Nodes[nodeIndex].ExtraPortMappings,
			portMapping,
		)

		return
	}

	return
}
