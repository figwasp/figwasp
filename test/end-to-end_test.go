package test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/joel-ling/alduin/test/pkg/clients"
	"github.com/joel-ling/alduin/test/pkg/clusters"
	"github.com/joel-ling/alduin/test/pkg/credentials"
	"github.com/joel-ling/alduin/test/pkg/deployments"
	"github.com/joel-ling/alduin/test/pkg/images"
	"github.com/joel-ling/alduin/test/pkg/permissions"
	"github.com/joel-ling/alduin/test/pkg/repositories"
	"github.com/joel-ling/alduin/test/pkg/secrets"
)

// As a Kubernetes administrator deploying container applications to a cluster,
// I want a rolling restart of a deployment to be automatically triggered
// whenever the tag of a currently-deployed image is inherited by another image
// so that the deployment is always up-to-date without manual intervention.

func TestEndToEnd(t *testing.T) {
	var (
		e error
	)

	// Given there is a container image repository

	const (
		dockerHost = "172.17.0.1"
		localhost  = "127.0.0.1"

		repositoryPort = 5000

		username = "username"
		password = "password"
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
		repositories.WithBasicAuthentication(username, password),
		repositories.WithTransportLayerSecurity(
			credential.PathToCertPEM(),
			credential.PathToKeyPEM(),
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	// And in the repository there is an image of a container
	// And the container is a HTTP server with a GET endpoint
	// And the endpoint responds to requests with a fixed HTTP status code
	// And the status code is preset via a container image build argument

	const (
		buildContextPath = ".."
		dockerfilePath0  = "test/build/http-status-code-server/Dockerfile"
		// relative to build context

		imageName0     = "http-status-code-server"
		imageRefFormat = "%s/%s"

		serverPort    = 8000
		serverPortKey = "SERVER_PORT"

		statusCode0   = http.StatusNoContent
		statusCodeKey = "STATUS_CODE"
	)

	var (
		image0                 *images.DockerImage
		imageRef0              string
		repositoryAddressLocal net.TCPAddr
	)

	repositoryAddressLocal = net.TCPAddr{
		IP:   net.ParseIP(localhost),
		Port: repositoryPort,
	}

	imageRef0 = fmt.Sprintf(imageRefFormat,
		repositoryAddressLocal.String(),
		imageName0,
	)

	image0, e = images.NewDockerImage(buildContextPath, dockerfilePath0,
		images.WithTag(imageRef0),
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

	e = image0.PushWithBasicAuth(os.Stderr, username, password)
	if e != nil {
		t.Error(e)
	}

	// And there is a Kubernetes cluster

	const (
		caCertsDir   = "/etc/ssl/certs/test.pem" // kindest/node based on Ubuntu
		clusterName  = "end-to-end-test-cluster"
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

	// And the server is deployed to the cluster using a Kubernetes deployment
	// And the image of the server is pulled from the repository
	// And the endpoint is exposed using a Kubernetes service

	const (
		secretName = "docker-registry-secret"

		deploymentLabelKey = "app"
	)

	var (
		repositoryAddressDocker net.TCPAddr

		secret *secrets.KubernetesDockerRegistrySecret

		deployment0 *deployments.KubernetesDeployment
	)

	repositoryAddressDocker = net.TCPAddr{
		IP:   net.ParseIP(dockerHost),
		Port: repositoryPort,
	}

	secret, e = secrets.NewKubernetesDockerRegistrySecret(
		cluster.KubeconfigPath(),
		secretName,
		repositoryAddressDocker.String(),
		username,
		password,
	)
	if e != nil {
		t.Error(e)
	}

	defer secret.Destroy()

	deployment0, e = deployments.NewKubernetesDeployment(
		imageName0,
		cluster.KubeconfigPath(),
		deployments.WithLabel(deploymentLabelKey, imageName0),
		deployments.WithContainerWithTCPPorts(
			imageName0,
			strings.ReplaceAll(imageRef0, localhost, dockerHost),
			serverPort,
		),
		deployments.WithImagePullSecrets(secretName),
		deployments.WithHostNetwork(),
	)
	if e != nil {
		t.Error(e)
	}

	defer deployment0.Destroy()

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

	// And Alduin is running in the cluster
	// And Alduin is authenticated as a Kubernetes service account
	// And the service account is authorised to get and patch deployments

	const (
		dockerfilePath1 = ""

		imageName1 = "alduin"
	)

	var (
		image1    *images.DockerImage
		imageRef1 string
	)

	imageRef1 = fmt.Sprintf(imageRefFormat,
		repositoryAddressLocal.String(),
		imageName1,
	)

	image1, e = images.NewDockerImage(buildContextPath, dockerfilePath1,
		images.WithTag(imageRef1),
		images.WithOutputStream(os.Stderr),
	)
	if e != nil {
		t.Error(e)
	}

	defer image1.Destroy()

	e = image1.PushWithBasicAuth(os.Stderr, username, password)
	if e != nil {
		t.Error(e)
	}

	const (
		resource = "deployments"
		verb0    = "get"
		verb1    = "patch"
	)

	var (
		permission *permissions.KubernetesRole
	)

	permission, e = permissions.NewKubernetesRole(
		imageName1,
		cluster.KubeconfigPath(),
		permissions.WithPolicyRule(
			[]string{verb0, verb1},
			[]string{resource},
		),
	)
	if e != nil {
		t.Error(e)
	}

	defer permission.Destroy()

	var (
		deployment1 *deployments.KubernetesDeployment
	)

	deployment1, e = deployments.NewKubernetesDeployment(
		imageName1,
		cluster.KubeconfigPath(),
		deployments.WithLabel(deploymentLabelKey, imageName1),
		deployments.WithContainerWithTCPPorts(
			imageName1,
			strings.ReplaceAll(imageRef1, localhost, dockerHost),
		),
		deployments.WithImagePullSecrets(secretName),
		deployments.WithServiceAccount(imageName1),
		deployments.WithHostNetwork(),
	)
	if e != nil {
		t.Error(e)
	}

	defer deployment1.Destroy()

	// When I rebuild the image so that it returns a different status code
	// And I transfer to the new image the tag of the existing image
	// And I push the new image to the repository

	const (
		statusCode1 = http.StatusTeapot
	)

	var (
		image2 *images.DockerImage
	)

	image2, e = images.NewDockerImage(buildContextPath, dockerfilePath0,
		images.WithTag(imageRef0),
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

	defer image2.Destroy()

	e = image2.PushWithBasicAuth(os.Stderr, username, password)
	if e != nil {
		t.Error(e)
	}

	// And I allow time for a rolling restart of the deployment to complete
	// And I send a request to the endpoint
	// Then I should see in the response to the request the new status code

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
