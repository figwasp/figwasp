package images

import (
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/joel-ling/alduin/pkg/images"
	"github.com/joel-ling/alduin/pkg/repositories"
)

func TestDockerImage(t *testing.T) {
	const (
		buildContextPath = "../.."
		dockerfilePath   = "cmd/http-status-code-server/Dockerfile"
		imageTag         = "%s/http-status-code-server"
		repositoryHost   = "127.0.0.1"
		repositoryPort   = 5000
		serverPortKey    = "SERVER_PORT"
		serverPortValue  = "8000"
		statusCodeKey    = "STATUS_CODE"
		statusCodeValue  = "418"
	)

	var (
		address    net.TCPAddr
		image      *images.DockerImage
		imageRef   string
		repository *repositories.DockerRegistry

		e error
	)

	address = net.TCPAddr{
		IP:   net.ParseIP(repositoryHost),
		Port: repositoryPort,
	}

	repository, e = repositories.NewDockerRegistry(address)
	if e != nil {
		t.Error(e)
	}

	defer repository.Close()

	image, e = images.NewDockerImage(buildContextPath, dockerfilePath)
	if e != nil {
		t.Error(e)
	}

	imageRef = fmt.Sprintf(imageTag,
		address.String(),
	)

	image.SetTag(imageRef)

	image.SetBuildArg(serverPortKey, serverPortValue)
	image.SetBuildArg(statusCodeKey, statusCodeValue)

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
}
