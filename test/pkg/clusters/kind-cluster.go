package clusters

import (
	_ "embed"
	"fmt"
	"net"

	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"
)

var (
	//go:embed containerd-config-patch.toml
	containerdConfigPatchTemplate string
)

type KindCluster struct {
	provider       *cluster.Provider
	nodeImageRef   string
	name           string
	config         *v1alpha4.Cluster
	kubeConfigPath string
}

func NewKindCluster(nodeImageRef, name, kubeConfigPath string) (
	c *KindCluster, e error,
) {
	var (
		logger         log.Logger
		providerOption cluster.ProviderOption
	)

	logger = cmd.NewLogger()

	providerOption = cluster.ProviderWithLogger(logger)

	c = &KindCluster{
		provider:     cluster.NewProvider(providerOption),
		nodeImageRef: nodeImageRef,
		name:         name,
		config: &v1alpha4.Cluster{
			Nodes: []v1alpha4.Node{
				{
					Role:              v1alpha4.ControlPlaneRole,
					ExtraPortMappings: make([]v1alpha4.PortMapping, 0),
				},
			},
			ContainerdConfigPatches: make([]string, 0),
		},
		kubeConfigPath: kubeConfigPath,
	}

	return
}

func (c *KindCluster) AddPortMapping(port int32) {
	// https://kind.sigs.k8s.io/docs/user/quick-start/
	//  #mapping-ports-to-the-host-machine

	const (
		controlPlaneNodeIndex = 0
	)

	c.config.Nodes[controlPlaneNodeIndex].ExtraPortMappings = append(
		c.config.Nodes[controlPlaneNodeIndex].ExtraPortMappings,
		v1alpha4.PortMapping{
			ContainerPort: port,
			HostPort:      port,
		},
	)

	return
}

func (c *KindCluster) AddHTTPRegistryMirror(address net.TCPAddr) {
	var (
		patch string
	)

	patch = fmt.Sprintf(containerdConfigPatchTemplate,
		address.String(),
		address.String(),
	)

	c.config.ContainerdConfigPatches = append(c.config.ContainerdConfigPatches,
		patch,
	)

	return
}

func (c *KindCluster) Create() (e error) {
	e = c.provider.Create(c.name,
		cluster.CreateWithNodeImage(c.nodeImageRef),
		cluster.CreateWithKubeconfigPath(c.kubeConfigPath),
		cluster.CreateWithV1Alpha4Config(c.config),
	)
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
