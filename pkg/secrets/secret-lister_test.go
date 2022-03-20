package secrets

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/joel-ling/alduin/test/pkg/clusters"
	"github.com/joel-ling/alduin/test/pkg/secrets"
)

func TestSecretLister(t *testing.T) {
	const (
		clusterName  = "test-secret-lister-cluster"
		nodeImageRef = "kindest/node:v1.23.3"

		secretName0 = "test-secret-0"
		secretName1 = "test-secret-1"

		registryAddress = "docker.io"

		username = "username"
		password = "password"

		masterURL = ""

		nSecrets = 2
	)

	var (
		cluster *clusters.KindCluster

		secret0 *secrets.KubernetesDockerRegistrySecret
		secret1 *secrets.KubernetesDockerRegistrySecret

		config *rest.Config
		lister SecretLister

		list []v1.Secret

		e error
	)

	cluster, e = clusters.NewKindCluster(nodeImageRef, clusterName)
	if e != nil {
		t.Error(e)
	}

	defer cluster.Destroy()

	secret0, e = secrets.NewKubernetesDockerRegistrySecret(
		cluster.KubeconfigPath(),
		secretName0,
		registryAddress,
		username,
		password,
	)
	if e != nil {
		t.Error(e)
	}

	defer secret0.Destroy()

	secret1, e = secrets.NewKubernetesDockerRegistrySecret(
		cluster.KubeconfigPath(),
		secretName1,
		registryAddress,
		username,
		password,
	)
	if e != nil {
		t.Error(e)
	}

	defer secret1.Destroy()

	config, e = clientcmd.BuildConfigFromFlags(
		masterURL,
		cluster.KubeconfigPath(),
	)
	if e != nil {
		t.Error(e)
	}

	lister, e = NewSecretLister(config, v1.NamespaceDefault)
	if e != nil {
		t.Error(e)
	}

	list, e = lister.ListSecrets(
		context.Background(),
	)
	if e != nil {
		t.Error(e)
	}

	assert.EqualValues(t,
		nSecrets,
		len(list),
	)
}
