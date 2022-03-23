package test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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

const (
	dockerHost = "172.17.0.1"
	localhost  = "127.0.0.1"

	buildContextPath = ".."
	imageRefFormat   = "%s/%s:latest"

	serverPort = 30000

	deploymentLabelKey = "app"
)

func TestEndToEnd(t *testing.T) {
	var (
		e error
	)

	// Given there is a container image repository

	var (
		repository *containerImageRepository
	)

	repository, e = thereIsAContainerImageRepository()
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	// And in the repository there is an image of a container
	// And the container is a HTTP server with a GET endpoint
	// And the endpoint responds to requests with a fixed HTTP status code
	// And the status code is preset via a container image build argument

	const (
		statusCode0 = http.StatusNoContent
	)

	var (
		image *containerImageOfAServer
	)

	image, e = thereIsAContainerImageOfAServer(repository, statusCode0)
	if e != nil {
		t.Error(e)
	}

	// And there is a Kubernetes cluster

	var (
		cluster *kubernetesCluster
	)

	cluster, e = thereIsAKubernetesCluster(repository)
	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	// And the server is deployed to the cluster using a Kubernetes deployment
	// And the image of the server is pulled from the repository
	// And the endpoint is exposed using a Kubernetes service

	var (
		deployment *deployments.KubernetesDeployment
	)

	deployment, e = thereIsAKubernetesDeployment(cluster, image)
	if e != nil {
		t.Error(e)
	}

	defer deployment.Destroy()

	/// TODO ///
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
	/// TODO ///

	// And Alduin is running in the cluster
	// And Alduin is authenticated as a Kubernetes service account
	// And the service account is authorised to get and patch deployments

	var (
		alduin *alduinInAPod
	)

	alduin, e = alduinIsRunningInAPod(cluster, repository)
	if e != nil {
		t.Error(e)
	}

	defer alduin.Destroy()

	// When I rebuild the image so that it returns a different status code
	// And I transfer to the new image the tag of the existing image
	// And I push the new image to the repository

	const (
		statusCode1 = http.StatusTeapot
	)

	_, e = thereIsAContainerImageOfAServer(repository, statusCode1)
	if e != nil {
		t.Error(e)
	}

	// And I allow time for a rolling restart of the deployment to complete
	// And I send a request to the endpoint
	// Then I should see in the response to the request the new status code

	const (
		timeout1 = 3 * time.Minute
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

type containerImageRepository struct {
	credential *credentials.TLSCertificate
	repository *repositories.DockerRegistry

	addressListen net.TCPAddr
	addressDocker net.TCPAddr
	addressLocal  net.TCPAddr
}

func thereIsAContainerImageRepository() (r *containerImageRepository, e error) {
	const (
		port = 5000
	)

	r = &containerImageRepository{
		addressListen: net.TCPAddr{
			Port: port,
		},
		addressDocker: net.TCPAddr{
			IP:   net.ParseIP(dockerHost),
			Port: port,
		},
		addressLocal: net.TCPAddr{
			IP:   net.ParseIP(localhost),
			Port: port,
		},
	}

	r.credential, e = credentials.NewTLSCertificate(
		credentials.WithExtendedKeyUsageForServerAuth(),
		credentials.WithIPAddress(localhost),
		credentials.WithIPAddress(dockerHost),
	)
	if e != nil {
		return
	}

	r.repository, e = repositories.NewDockerRegistry(
		r.addressListen,
		repositories.WithBasicAuthentication(
			r.Username(),
			r.Password(),
		),
		repositories.WithTransportLayerSecurity(
			r.credential.PathToCertPEM(),
			r.credential.PathToKeyPEM(),
		),
	)
	if e != nil {
		return
	}

	return
}

func (r *containerImageRepository) AddressDocker() string {
	return r.addressDocker.String()
}

func (r *containerImageRepository) AddressLocal() string {
	return r.addressLocal.String()
}

func (r *containerImageRepository) Credential() *credentials.TLSCertificate {
	return r.credential
}

func (r *containerImageRepository) Username() string {
	return "username"
}

func (r *containerImageRepository) Password() string {
	return "password"
}

func (r *containerImageRepository) Destroy() (e error) {
	e = r.credential.Destroy()
	if e != nil {
		return
	}

	e = r.repository.Destroy()
	if e != nil {
		return
	}

	return
}

type containerImage struct {
	imageRefDocker string
	imageRefLocal  string
}

func (i *containerImage) ImageRefDocker() string {
	return i.imageRefDocker
}

func (i *containerImage) ImageRefLocal() string {
	return i.imageRefLocal
}

type containerImageOfAServer struct {
	containerImage
	statusCode int
}

func thereIsAContainerImageOfAServer(
	repository *containerImageRepository, statusCode int,
) (
	i *containerImageOfAServer, e error,
) {
	const (
		dockerfilePath = "test/build/http-status-code-server/Dockerfile"
		// relative to build context

		serverPortKey = "SERVER_PORT"
		statusCodeKey = "STATUS_CODE"
	)

	var (
		image *images.DockerImage
	)

	i = &containerImageOfAServer{
		containerImage: containerImage{
			imageRefDocker: fmt.Sprintf(imageRefFormat,
				repository.AddressDocker(),
				i.ImageName(),
			),
			imageRefLocal: fmt.Sprintf(imageRefFormat,
				repository.AddressLocal(),
				i.ImageName(),
			),
		},
	}

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath,
		images.WithTag(i.imageRefLocal),
		images.WithBuildArg(serverPortKey,
			fmt.Sprint(serverPort),
		),
		images.WithBuildArg(statusCodeKey,
			fmt.Sprint(statusCode),
		),
		images.WithOutputStream(os.Stderr),
	)
	if e != nil {
		return
	}

	e = image.PushWithBasicAuth(
		os.Stderr,
		repository.Username(),
		repository.Password(),
	)
	if e != nil {
		return
	}

	e = image.Destroy()
	if e != nil {
		return
	}

	return
}

func (i *containerImageOfAServer) ImageName() string {
	return "http-status-code-server"
}

func (i *containerImageOfAServer) StatusCode() int {
	return i.statusCode
}

type containerImageOfAlduin struct {
	containerImage
}

func thereIsAContainerImageOfAlduin(repository *containerImageRepository) (
	i *containerImageOfAlduin, e error,
) {
	const (
		commandArg0    = "-c"
		commandArg1    = "CGO_ENABLED=0 GOOS=linux go build -o bin/alduin ../cmd/alduin"
		commandName    = "bash"
		dockerfilePath = "test/build/alduin/Dockerfile"
	)

	var (
		command *exec.Cmd
		image   *images.DockerImage
	)

	command = exec.Command(commandName, commandArg0, commandArg1)

	e = command.Run()
	if e != nil {
		return
	}

	i = &containerImageOfAlduin{
		containerImage: containerImage{
			imageRefDocker: fmt.Sprintf(imageRefFormat,
				repository.AddressDocker(),
				i.ImageName(),
			),
			imageRefLocal: fmt.Sprintf(imageRefFormat,
				repository.AddressLocal(),
				i.ImageName(),
			),
		},
	}

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath,
		images.WithTag(i.imageRefLocal),
		images.WithOutputStream(os.Stderr),
	)
	if e != nil {
		return
	}

	e = image.PushWithBasicAuth(
		os.Stderr,
		repository.Username(),
		repository.Password(),
	)
	if e != nil {
		return
	}

	e = image.Destroy()
	if e != nil {
		return
	}

	return
}

func (i *containerImageOfAlduin) ImageName() string {
	return "alduin"
}

type kubernetesCluster struct {
	cluster *clusters.KindCluster
	secret  *secrets.KubernetesDockerRegistrySecret
}

func thereIsAKubernetesCluster(repository *containerImageRepository) (
	c *kubernetesCluster, e error,
) {
	const (
		clusterName  = "end-to-end-test-cluster"
		nodeImageRef = "kindest/node:v1.23.3"
	)

	c = &kubernetesCluster{}

	c.cluster, e = clusters.NewKindCluster(nodeImageRef, clusterName,
		clusters.WithExtraMounts(
			c.NodeCACertsDir(),
			repository.Credential().PathToCertPEM(),
		),
		clusters.WithExtraPortMapping(
			serverPort,
		),
	)
	if e != nil {
		return
	}

	c.secret, e = secrets.NewKubernetesDockerRegistrySecret(
		c.cluster.KubeconfigPath(),
		c.DockerRegistrySecretName(),
		repository.AddressDocker(),
		repository.Username(),
		repository.Password(),
	)
	if e != nil {
		return
	}

	return
}

func (c *kubernetesCluster) DockerRegistrySecretName() string {
	return "docker-registry-secret"
}

func (c *kubernetesCluster) KubeconfigPath() string {
	return c.cluster.KubeconfigPath()
}

func (c *kubernetesCluster) NodeCACertsDir() string {
	return "/etc/ssl/certs/test.pem" // image "kindest/node" is based on Ubuntu
}

func (c *kubernetesCluster) Destroy() (e error) {
	e = c.secret.Destroy()
	if e != nil {
		return
	}

	e = c.cluster.Destroy()
	if e != nil {
		return
	}

	return
}

func thereIsAKubernetesDeployment(
	cluster *kubernetesCluster, image *containerImageOfAServer,
) (
	d *deployments.KubernetesDeployment, e error,
) {
	d, e = deployments.NewKubernetesDeployment(
		image.ImageName(),
		cluster.KubeconfigPath(),
		deployments.WithLabel(
			deploymentLabelKey,
			image.ImageName(),
		),
		deployments.WithContainerWithTCPPorts(
			image.ImageName(),
			image.ImageRefDocker(),
			serverPort,
		),
		deployments.WithImagePullSecrets(
			cluster.DockerRegistrySecretName(),
		),
	)
	if e != nil {
		return
	}

	return
}

type alduinInAPod struct {
	deployment *deployments.KubernetesDeployment
	permission *permissions.KubernetesRole
}

func alduinIsRunningInAPod(
	cluster *kubernetesCluster, repository *containerImageRepository,
) (
	a *alduinInAPod, e error,
) {
	const (
		apiGroup0 = ""
		apiGroup1 = "apps"
		resource0 = "deployments"
		resource1 = "replicasets"
		resource2 = "pods"
		resource3 = "secrets"
		verb0     = "get"
		verb1     = "update"
		verb2     = "list"

		volumeName = "ca-certs"
	)

	var (
		image *containerImageOfAlduin
	)

	a = &alduinInAPod{}

	image, e = thereIsAContainerImageOfAlduin(repository)
	if e != nil {
		return
	}

	a.permission, e = permissions.NewKubernetesRole(
		image.ImageName(),
		cluster.KubeconfigPath(),
		permissions.WithPolicyRule(
			[]string{verb0, verb1},
			[]string{apiGroup1},
			[]string{resource0},
		),
		permissions.WithPolicyRule(
			[]string{verb2},
			[]string{apiGroup1},
			[]string{resource1},
		),
		permissions.WithPolicyRule(
			[]string{verb2},
			[]string{apiGroup0},
			[]string{resource2},
		),
		permissions.WithPolicyRule(
			[]string{verb2},
			[]string{apiGroup0},
			[]string{resource3},
		),
	)
	if e != nil {
		return
	}

	a.deployment, e = deployments.NewKubernetesDeployment(
		image.ImageName(),
		cluster.KubeconfigPath(),
		deployments.WithLabel(
			deploymentLabelKey,
			image.ImageName(),
		),
		deployments.WithContainerWithTCPPorts(
			image.ImageName(),
			image.ImageRefDocker(),
		),
		deployments.WithImagePullSecrets(
			cluster.DockerRegistrySecretName(),
		),
		deployments.WithServiceAccount(
			image.ImageName(),
		),
		deployments.WithHostPathVolume(
			volumeName,
			cluster.NodeCACertsDir(),
			cluster.NodeCACertsDir(),
			image.ImageName(),
		),
	)
	if e != nil {
		return
	}

	return
}

func (a *alduinInAPod) Destroy() (e error) {
	e = a.deployment.Destroy()
	if e != nil {
		return
	}

	e = a.permission.Destroy()
	if e != nil {
		return
	}

	return
}
