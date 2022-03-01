package selectors

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/joel-ling/alduin/test/pkg/clusters"
	"github.com/joel-ling/alduin/test/pkg/deployments"
	"github.com/joel-ling/alduin/test/pkg/images"
	"github.com/joel-ling/alduin/test/pkg/repositories"
)

func TestDeploymentPodSelector(t *testing.T) {
	const (
		repositoryPort = 5000

		buildContextPath = "../.."
		dockerfilePath   = "test/build/idle/Dockerfile"
		imageName        = "idle"
		imageRefFormat   = "%s/%s"
		localhost        = "127.0.0.1"

		clusterName  = "test-cluster"
		dockerHost   = "172.17.0.1"
		masterURL    = ""
		nodeImageRef = "kindest/node:v1.23.3"

		deploymentLabelKey = "app"
		deploymentName     = "test-deployment"
		serviceAccountName = ""

		nPods = 1
	)

	var (
		repositoryAddress net.TCPAddr

		image    *images.DockerImage
		imageRef string

		cluster *clusters.KindCluster

		deployment *deployments.KubernetesDeployment

		config   *rest.Config
		selector PodSelector

		pods []v1.Pod

		e error
	)

	repositoryAddress = net.TCPAddr{
		Port: repositoryPort,
	}

	_, e = repositories.NewDockerRegistry(repositoryAddress)
	if e != nil {
		t.Error(e)
	}

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath)
	if e != nil {
		t.Error(e)
	}

	repositoryAddress.IP = net.ParseIP(localhost)

	imageRef = fmt.Sprintf(imageRefFormat,
		repositoryAddress.String(),
		imageName,
	)

	image.SetTag(imageRef)

	e = image.Build(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	e = image.Push(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	cluster, e = clusters.NewKindCluster(
		nodeImageRef,
		clusterName,
	)
	if e != nil {
		t.Error(e)
	}

	repositoryAddress.IP = net.ParseIP(dockerHost)

	cluster.AddHTTPRegistryMirror(repositoryAddress)

	e = cluster.Create()
	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	deployment, e = deployments.NewKubernetesDeployment(
		deploymentName,
		serviceAccountName,
		cluster.KubeconfigPath(),
	)
	if e != nil {
		t.Error(e)
	}

	deployment.SetLabel(deploymentLabelKey, imageName)

	imageRef = fmt.Sprintf(imageRefFormat,
		repositoryAddress.String(),
		imageName,
	)

	deployment.AddContainerWithoutPorts(imageName, imageRef)

	e = deployment.Create()
	if e != nil {
		t.Error(e)
	}

	defer deployment.Delete()

	config, e = clientcmd.BuildConfigFromFlags(
		masterURL,
		cluster.KubeconfigPath(),
	)
	if e != nil {
		t.Error(e)
	}

	selector, e = NewDeploymentPodSelector(
		config,
		v1.NamespaceDefault,
	)
	if e != nil {
		t.Error(e)
	}

	pods, e = selector.SelectPods(deploymentName,
		context.Background(),
	)
	if e != nil {
		t.Error(e)
	}

	assert.Equal(t,
		nPods,
		len(pods),
	)
}
