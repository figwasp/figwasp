package clusters

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"net"
	"os"

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
	kubeconfigPath string
}

func NewKindCluster(nodeImageRef, name string) (
	c *KindCluster, e error,
) {
	const (
		// See https://pkg.go.dev/io/ioutil#TempFile
		kubeconfigDirectory = ""
		kubeconfigFilename  = "*"
	)

	var (
		kubeconfigFile *os.File
		logger         log.Logger
		providerOption cluster.ProviderOption
	)

	kubeconfigFile, e = ioutil.TempFile(
		kubeconfigDirectory,
		kubeconfigFilename,
	)
	if e != nil {
		return
	}

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
		kubeconfigPath: kubeconfigFile.Name(),
	}

	return
}

func (c *KindCluster) KubeconfigPath() string {
	return c.kubeconfigPath
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
		cluster.CreateWithKubeconfigPath(c.kubeconfigPath),
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
