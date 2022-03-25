package figwasp

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/figwasp/figwasp/test/pkg/clusters"
	creds "github.com/figwasp/figwasp/test/pkg/credentials"
	"github.com/figwasp/figwasp/test/pkg/deployments"
	"github.com/figwasp/figwasp/test/pkg/images"
	"github.com/figwasp/figwasp/test/pkg/repositories"
)

func TestPodLister(t *testing.T) {
	var (
		e error
	)

	const (
		dockerHost = "172.17.0.1"
		localhost  = "127.0.0.1"

		repositoryPort = 5001
	)

	var (
		credential *creds.TLSCertificate

		repository        *repositories.DockerRegistry
		repositoryAddress net.TCPAddr
	)

	credential, e = creds.NewTLSCertificate(
		creds.WithExtendedKeyUsageForServerAuth(),
		creds.WithIPAddress(localhost),
		creds.WithIPAddress(dockerHost),
	)
	if e != nil {
		t.Error(e)
	}

	defer credential.Destroy()

	repositoryAddress = net.TCPAddr{
		Port: repositoryPort,
	}

	repository, e = repositories.NewDockerRegistry(repositoryAddress,
		repositories.WithTransportLayerSecurity(
			credential.PathToCertPEM(),
			credential.PathToKeyPEM(),
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	const (
		buildContextPath = "../.."
		dockerfilePath   = "test/build/idle/Dockerfile"
		// relative to build context

		imageName      = "idle"
		imageRefFormat = "%s/%s"
	)

	var (
		image                  *images.DockerImage
		imageRef               string
		repositoryAddressLocal net.TCPAddr
	)

	repositoryAddressLocal = net.TCPAddr{
		IP:   net.ParseIP(localhost),
		Port: repositoryPort,
	}

	imageRef = fmt.Sprintf(imageRefFormat,
		repositoryAddressLocal.String(),
		imageName,
	)

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath,
		images.WithTag(imageRef),
		images.WithOutputStream(os.Stderr),
	)
	if e != nil {
		t.Error(e)
	}

	defer image.Destroy()

	e = image.Push(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	const (
		caCertsDir   = "/etc/ssl/certs/test.pem" // kindest/node based on Ubuntu
		clusterName  = "test-pod-lister-cluster"
		nodeImageRef = "kindest/node:v1.23.3"
	)

	var (
		cluster *clusters.KindCluster
	)

	cluster, e = clusters.NewKindCluster(nodeImageRef, clusterName,
		clusters.WithExtraMounts(
			caCertsDir,
			credential.PathToCertPEM(),
		),
	)

	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	const (
		deploymentLabelKey = "app"
		deploymentName0    = "deployment0"
		deploymentName1    = "deployment1"
	)

	var (
		deployment0 *deployments.KubernetesDeployment
		deployment1 *deployments.KubernetesDeployment
	)

	deployment0, e = deployments.NewKubernetesDeployment(
		deploymentName0,
		cluster.KubeconfigPath(),
		deployments.WithLabel(deploymentLabelKey, deploymentName0),
		deployments.WithContainerWithTCPPorts(imageName,
			strings.ReplaceAll(imageRef, localhost, dockerHost),
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer deployment0.Destroy()

	deployment1, e = deployments.NewKubernetesDeployment(
		deploymentName1,
		cluster.KubeconfigPath(),
		deployments.WithLabel(deploymentLabelKey, deploymentName1),
		deployments.WithContainerWithTCPPorts(imageName,
			strings.ReplaceAll(imageRef, localhost, dockerHost),
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer deployment1.Destroy()

	const (
		masterURL = ""

		nPods = 1
	)

	var (
		config *rest.Config
		lister *deploymentPodLister
		pods   []v1.Pod
	)

	config, e = clientcmd.BuildConfigFromFlags(
		masterURL,
		cluster.KubeconfigPath(),
	)
	if e != nil {
		t.Error(e)
	}

	lister, e = NewDeploymentPodLister(config, v1.NamespaceDefault)
	if e != nil {
		t.Error(e)
	}

	pods, e = lister.ListPods(deploymentName0,
		context.Background(),
	)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t,
		nPods,
		len(pods),
	)

	pods, e = lister.ListPods(deploymentName1,
		context.Background(),
	)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t,
		nPods,
		len(pods),
	)
}
