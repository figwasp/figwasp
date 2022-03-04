package secrets

import (
	"context"
	"encoding/json"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/create"
)

type KubernetesDockerRegistrySecret struct {
	secrets typedCoreV1.SecretInterface
	secret  *coreV1.Secret
}

func NewKubernetesDockerRegistrySecret(
	kubeconfigPath, secretName, registryAddress, username, password string,
) (
	s *KubernetesDockerRegistrySecret, e error,
) {
	const (
		masterURL = ""
	)

	var (
		clientset *kubernetes.Clientset
		config    *rest.Config

		dockerConfigJSON create.DockerConfigJSON
	)

	config, e = clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if e != nil {
		return
	}

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		return
	}

	s = &KubernetesDockerRegistrySecret{
		secrets: clientset.CoreV1().Secrets(
			coreV1.NamespaceDefault,
		),
		secret: &coreV1.Secret{
			ObjectMeta: metaV1.ObjectMeta{
				Name: secretName,
			},
			Data: make(map[string][]byte),
			Type: coreV1.SecretTypeDockerConfigJson,
		},
	}

	dockerConfigJSON.Auths = map[string]create.DockerConfigEntry{
		registryAddress: {
			Username: username,
			Password: password,
		},
	}

	s.secret.Data[coreV1.DockerConfigJsonKey], e = json.Marshal(
		dockerConfigJSON,
	)
	if e != nil {
		return
	}

	s.secret, e = s.secrets.Create(
		context.Background(),
		s.secret,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	return
}

func (s *KubernetesDockerRegistrySecret) Destroy() (e error) {
	e = s.secrets.Delete(
		context.Background(),
		s.secret.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{},
	)
	if e != nil {
		return
	}

	return
}
