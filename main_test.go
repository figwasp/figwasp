package main

import (
	"time"

	"github.com/joel-ling/alduin/test"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	const (
		status0 = 418
		status1 = 204
		timeout = time.Second
	)

	var (
		client     test.Client
		cluster    test.Cluster
		repository test.Repository

		status int

		e error
	)

	repository, e = test.NewRepository()
	if e != nil {
		t.Error(e)
	}

	defer repository.Destroy()

	e = repository.BuildAndPushServerImage(status0)
	if e != nil {
		t.Error(e)
	}

	e = repository.BuildAndPushAlduinImage()
	if e != nil {
		t.Error(e)
	}

	cluster, e = test.NewCluster()
	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	e = cluster.DeployServer()
	if e != nil {
		t.Error(e)
	}

	e = cluster.DeployAlduin()
	if e != nil {
		t.Error(e)
	}

	client, e = test.NewClient()
	if e != nil {
		t.Error(e)
	}

	status, e = client.SendRequestToServerEndpoint(timeout)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, status0, status)

	e = repository.BuildAndPushServerImage(status1)
	if e != nil {
		t.Error(e)
	}

	status, e = client.SendRequestToServerEndpoint(timeout)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t, status1, status)
}
