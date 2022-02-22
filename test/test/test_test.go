package test

import (
	"fmt"
	"io/ioutil"
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
	"github.com/joel-ling/alduin/test/pkg/deployments"
	"github.com/joel-ling/alduin/test/pkg/images"
	"github.com/joel-ling/alduin/test/pkg/repositories"
)

func TestTest(t *testing.T) {
	const (
		any            = "0.0.0.0"
		dockerHost     = "172.17.0.1"
		localhost      = "127.0.0.1"
		repositoryPort = 5000

		clusterName         = "test-cluster"
		kubeConfigDirectory = "/tmp"
		kubeConfigPattern   = "*"
		nodeImageRef        = "kindest/node:v1.21.1"

		buildContextPath = "../.."
		dockerfilePath   = "test/build/http-status-code-server/Dockerfile"
		imageName        = "http-status-code-server"
		imageRefFormat   = "%s/%s"

		serverPortKey   = "SERVER_PORT"
		serverPortValue = 8000
		statusCodeKey   = "STATUS_CODE"
		statusCodeValue = http.StatusTeapot

		deploymentLabelKey   = "app"
		deploymentLabelValue = "test"
		deploymentName       = "test-deployment"
		serviceAccountName   = ""

		scheme  = "http"
		timeout = time.Second
	)

	var (
		cluster           *clusters.KindCluster
		kubeConfigFile    *os.File
		repositoryAddress net.TCPAddr

		repository *repositories.DockerRegistry

		image    *images.DockerImage
		imageRef string

		deployment *deployments.KubernetesDeployment

		client        *clients.HTTPClient
		endpoint      url.URL
		serverAddress net.TCPAddr

		status int

		e error
	)

	// start Kubernetes cluster

	kubeConfigFile, e = ioutil.TempFile(
		kubeConfigDirectory,
		kubeConfigPattern,
	)
	if e != nil {
		t.Error(e)
	}

	defer os.Remove(
		kubeConfigFile.Name(),
	)

	cluster, e = clusters.NewKindCluster(
		nodeImageRef,
		clusterName,
		kubeConfigFile.Name(),
	)
	if e != nil {
		t.Error(e)
	}

	cluster.AddPortMapping(serverPortValue)

	repositoryAddress = net.TCPAddr{
		IP:   net.ParseIP(dockerHost),
		Port: repositoryPort,
	}

	cluster.AddHTTPRegistryMirror(repositoryAddress)

	e = cluster.Create()
	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	// set up container image repository

	repositoryAddress = net.TCPAddr{
		IP:   net.ParseIP(any),
		Port: repositoryPort,
	}

	repository, e = repositories.NewDockerRegistry(repositoryAddress)
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	// build and push container image of server to repository

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath)
	if e != nil {
		t.Error(e)
	}

	repositoryAddress = net.TCPAddr{
		IP:   net.ParseIP(localhost),
		Port: repositoryPort,
	}

	imageRef = fmt.Sprintf(imageRefFormat,
		repositoryAddress.String(),
		imageName,
	)

	image.SetTag(imageRef)

	image.SetBuildArg(serverPortKey,
		fmt.Sprint(serverPortValue),
	)

	image.SetBuildArg(statusCodeKey,
		fmt.Sprint(statusCodeValue),
	)

	e = image.Build(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	e = image.Push(os.Stderr)
	if e != nil {
		t.Error(e)
	}

	e = image.Remove()
	if e != nil {
		t.Error(e)
	}

	// deploy test server to Kubernetes cluster

	deployment, e = deployments.NewKubernetesDeployment(
		deploymentName,
		serviceAccountName,
		kubeConfigFile.Name(),
	)
	if e != nil {
		t.Error(e)
	}

	deployment.SetLabel(deploymentLabelKey, deploymentLabelValue)

	deployment.AddContainerWithSingleTCPPort(
		imageName,
		strings.ReplaceAll(imageRef, localhost, dockerHost),
		serverPortValue,
	)

	e = deployment.Create()
	if e != nil {
		t.Error(e)
	}

	defer deployment.Delete()

	// interact with Kubernetes service via client

	client, e = clients.NewHTTPClient()
	if e != nil {
		t.Error(e)
	}

	serverAddress.Port = serverPortValue

	endpoint = url.URL{
		Scheme: scheme,
		Host:   serverAddress.String(),
	}

	status, e = client.GetStatusCodeFromEndpoint(endpoint, timeout)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, statusCodeValue, status)
}
