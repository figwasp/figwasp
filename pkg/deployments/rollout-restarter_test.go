package deployments

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/figwasp/figwasp/test/pkg/clients"
	"github.com/figwasp/figwasp/test/pkg/clusters"
	"github.com/figwasp/figwasp/test/pkg/credentials"
	"github.com/figwasp/figwasp/test/pkg/deployments"
	"github.com/figwasp/figwasp/test/pkg/images"
	"github.com/figwasp/figwasp/test/pkg/repositories"
)

func TestRolloutRestarter(t *testing.T) {
	var (
		e error
	)

	const (
		dockerHost = "172.17.0.1"
		localhost  = "127.0.0.1"

		repositoryPort = 5000
	)

	var (
		credential *credentials.TLSCertificate

		repository        *repositories.DockerRegistry
		repositoryAddress net.TCPAddr
	)

	credential, e = credentials.NewTLSCertificate(
		credentials.WithExtendedKeyUsageForServerAuth(),
		credentials.WithIPAddress(localhost),
		credentials.WithIPAddress(dockerHost),
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
		dockerfilePath   = "test/build/http-status-code-server/Dockerfile"
		// relative to build context

		imageName      = "http-status-code-server"
		imageRefFormat = "%s/%s"

		serverPort    = 30000
		serverPortKey = "SERVER_PORT"

		statusCode0   = http.StatusNoContent
		statusCodeKey = "STATUS_CODE"
	)

	var (
		image0                 *images.DockerImage
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

	image0, e = images.NewDockerImage(buildContextPath, dockerfilePath,
		images.WithTag(imageRef),
		images.WithBuildArg(serverPortKey,
			fmt.Sprint(serverPort),
		),
		images.WithBuildArg(statusCodeKey,
			fmt.Sprint(statusCode0),
		),
		images.WithOutputStream(os.Stderr),
	)
	if e != nil {
		t.Error(e)
	}

	defer image0.Destroy()

	e = image0.Push(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	const (
		caCertsDir   = "/etc/ssl/certs/test.pem" // kindest/node based on Ubuntu
		clusterName  = "test-rollout-restarter-cluster"
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
		clusters.WithExtraPortMapping(serverPort),
	)

	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	const (
		deploymentLabelKey = "app"
		deploymentName     = "deployment"
	)

	var (
		deployment *deployments.KubernetesDeployment
	)

	deployment, e = deployments.NewKubernetesDeployment(
		deploymentName,
		cluster.KubeconfigPath(),
		deployments.WithLabel(deploymentLabelKey, deploymentName),
		deployments.WithContainerWithTCPPorts(imageName,
			strings.ReplaceAll(imageRef, localhost, dockerHost),
			serverPort,
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer deployment.Destroy()

	const (
		scheme   = "http"
		timeout0 = time.Second
	)

	var (
		client        *clients.HTTPClient
		endpoint      url.URL
		serverAddress net.TCPAddr
		status        int
	)

	client, e = clients.NewHTTPClient()
	if e != nil {
		t.Error(e)
	}

	serverAddress.Port = serverPort

	endpoint = url.URL{
		Scheme: scheme,
		Host:   serverAddress.String(),
	}

	status, e = client.GetStatusCodeFromEndpoint(endpoint, timeout0)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, statusCode0, status)

	const (
		statusCode1 = http.StatusTeapot
	)

	var (
		image1 *images.DockerImage
	)

	image1, e = images.NewDockerImage(buildContextPath, dockerfilePath,
		images.WithTag(imageRef),
		images.WithBuildArg(serverPortKey,
			fmt.Sprint(serverPort),
		),
		images.WithBuildArg(statusCodeKey,
			fmt.Sprint(statusCode1),
		),
		images.WithOutputStream(os.Stderr),
	)
	if e != nil {
		t.Error(e)
	}

	defer image1.Destroy()

	e = image1.Push(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	const (
		masterURL = ""
	)

	var (
		config    *rest.Config
		restarter RolloutRestarter
	)

	config, e = clientcmd.BuildConfigFromFlags(
		masterURL,
		cluster.KubeconfigPath(),
	)
	if e != nil {
		t.Error(e)
	}

	restarter, e = NewDeploymentRolloutRestarter(config, v1.NamespaceDefault)
	if e != nil {
		t.Error(e)
	}

	e = restarter.RolloutRestart(deploymentName,
		context.Background(),
	)
	if e != nil {
		t.Error(e)
	}

	const (
		timeout1 = time.Minute
	)

	assert.Eventually(t,
		func() bool {
			status, e = client.GetStatusCodeFromEndpoint(endpoint, timeout0)
			if e != nil {
				t.Error(e)
			}

			return (status == statusCode1)
		},
		timeout1,
		timeout0,
	)
}
