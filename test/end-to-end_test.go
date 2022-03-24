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

	"github.com/figwasp/figwasp/test/pkg/clients"
	"github.com/figwasp/figwasp/test/pkg/clusters"
	"github.com/figwasp/figwasp/test/pkg/credentials"
	"github.com/figwasp/figwasp/test/pkg/deployments"
	"github.com/figwasp/figwasp/test/pkg/images"
	"github.com/figwasp/figwasp/test/pkg/jobs"
	"github.com/figwasp/figwasp/test/pkg/permissions"
	"github.com/figwasp/figwasp/test/pkg/repositories"
	"github.com/figwasp/figwasp/test/pkg/secrets"
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

	timeout0 = time.Second
	timeout1 = time.Minute
)

func TestEndToEnd(t *testing.T) {
	const (
		statusCode0 = http.StatusNoContent
		statusCode1 = http.StatusTeapot
	)

	var (
		figwasp    *figwaspInAPod
		client     *client
		cluster    *kubernetesCluster
		deployment *deployments.KubernetesDeployment
		image      *containerImageOfAServer
		repository *containerImageRepository

		status int

		e error
	)

	// Given there is a container image repository

	repository, e = thereIsAContainerImageRepository()
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	// And (in the repository) there is a container image of a server

	image, e = thereIsAContainerImageOfAServer(repository, statusCode0)
	if e != nil {
		t.Error(e)
	}

	// And there is a Kubernetes cluster

	cluster, e = thereIsAKubernetesCluster(repository)
	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	// And (in the cluster) there is a Kubernetes Deployment (of that server)

	deployment, e = thereIsAKubernetesDeployment(cluster, image)
	if e != nil {
		t.Error(e)
	}

	defer deployment.Destroy()

	// And there is a client (to the server obtaining its response to a request)

	client, e = thereIsAClient()
	if e != nil {
		t.Error(e)
	}

	status, e = client.ObtainServerResponse()
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, statusCode0, status)

	// When there is a (new) container image of [the] server (in the repository)

	_, e = thereIsAContainerImageOfAServer(repository, statusCode1)
	if e != nil {
		t.Error(e)
	}

	// And Figwasp is run as a Kubernetes Job (or CronJob, in the cluster)

	figwasp, e = figwaspIsRunningInAPod(cluster, repository)
	if e != nil {
		t.Error(e)
	}

	defer figwasp.Destroy()

	// Then the client should detect a corresponding change in server response

	assert.Eventually(t,
		func() bool {
			status, e = client.ObtainServerResponse()
			if e != nil {
				t.Error(e)
			}

			return (status == statusCode1)
		},
		timeout1,
		timeout0,
	)
}

type client struct {
	client   *clients.HTTPClient
	endpoint url.URL
}

func thereIsAClient() (c *client, e error) {
	const (
		scheme = "http"
	)

	var (
		serverAddress net.TCPAddr
	)

	serverAddress.Port = serverPort

	c = &client{
		endpoint: url.URL{
			Scheme: scheme,
			Host:   serverAddress.String(),
		},
	}

	c.client, e = clients.NewHTTPClient()
	if e != nil {
		return
	}

	return
}

func (c *client) ObtainServerResponse() (status int, e error) {
	status, e = c.client.GetStatusCodeFromEndpoint(c.endpoint, timeout0)
	if e != nil {
		return
	}

	return
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

type containerImageOfFigwasp struct {
	containerImage
}

func thereIsAContainerImageOfFigwasp(repository *containerImageRepository) (
	i *containerImageOfFigwasp, e error,
) {
	const (
		commandArg0    = "build"
		commandArg1    = "-o"
		commandArg2    = "bin/figwasp"
		commandArg3    = "../cmd/figwasp"
		commandEnv0    = "CGO_ENABLED=0"
		commandEnv1    = "GOOS=linux"
		commandName    = "go"
		dockerfilePath = "test/build/figwasp/Dockerfile"
	)

	var (
		command *exec.Cmd
		image   *images.DockerImage
	)

	command = exec.Command(commandName,
		commandArg0,
		commandArg1,
		commandArg2,
		commandArg3,
	)

	command.Env = append(os.Environ(),
		commandEnv0,
		commandEnv1,
	)

	e = command.Run()
	if e != nil {
		return
	}

	i = &containerImageOfFigwasp{
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

func (i *containerImageOfFigwasp) ImageName() string {
	return "figwasp"
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

type figwaspInAPod struct {
	job        *jobs.KubernetesJob
	permission *permissions.KubernetesRole
}

func figwaspIsRunningInAPod(
	cluster *kubernetesCluster, repository *containerImageRepository,
) (
	a *figwaspInAPod, e error,
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

		envVar = "FIGWASP_TARGET_DEPLOYMENT http-status-code-server"

		volumeName = "ca-certs"
	)

	var (
		image *containerImageOfFigwasp
	)

	a = &figwaspInAPod{}

	image, e = thereIsAContainerImageOfFigwasp(repository)
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

	a.job, e = jobs.NewKubernetesJob(
		image.ImageName(),
		cluster.KubeconfigPath(),
		jobs.WithLabel(
			deploymentLabelKey,
			image.ImageName(),
		),
		jobs.WithContainerWithEnvVars(
			image.ImageName(),
			image.ImageRefDocker(),
			envVar,
		),
		jobs.WithImagePullSecrets(
			cluster.DockerRegistrySecretName(),
		),
		jobs.WithServiceAccount(
			image.ImageName(),
		),
		jobs.WithHostPathVolume(
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

func (a *figwaspInAPod) Destroy() (e error) {
	e = a.job.Destroy()
	if e != nil {
		return
	}

	e = a.permission.Destroy()
	if e != nil {
		return
	}

	return
}
