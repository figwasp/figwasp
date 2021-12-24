package docker

import (
	"context"
	"testing"

	"github.com/joel-ling/lookout/docker/test"
)

func TestClient(t *testing.T) {
	const (
		address           = "localhost:5000"
		emptyString       = ""
		expectedImageHash = "sha256:" +
			"1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792"
		htpasswdPath = "test/htpasswd"
		imageNameTag = "localhost:5000/hello-world:latest"
		password     = "FcaVaAURjywghqVGHgYzbvviejUdTPTxKm53m-YLC10"
		storePath    = "test/registry"
		tlsCertPath  = "test/localhost.crt"
		tlsKeyPath   = "test/localhost.key"
		username     = "docker"
	)

	var (
		client    *Client
		ctx       context.Context = context.Background()
		e         error
		imageHash string
	)

	e = test.ServeRegistry(
		address,
		tlsCertPath,
		tlsKeyPath,
		htpasswdPath,
		storePath,
	)
	if e != nil {
		t.Error(e)
	}

	client, e = NewClient(ctx, address, username, password)
	if e != nil {
		t.Error(e)
	}

	imageHash, e = client.RetrieveImageHash(ctx, imageNameTag)
	if e != nil {
		t.Error(e)
	}

	if imageHash != expectedImageHash {
		t.Log()
		t.Fail()
	}
}
