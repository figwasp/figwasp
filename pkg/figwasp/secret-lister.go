package figwasp

import (
	"context"

	"github.com/juju/errors"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type secretLister struct {
	secrets typedCoreV1.SecretInterface
}

func NewSecretLister(config *rest.Config, namespace string) (
	l *secretLister, e error,
) {
	var (
		clientset *kubernetes.Clientset
	)

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	l = &secretLister{
		secrets: clientset.CoreV1().Secrets(namespace),
	}

	return
}

func (l *secretLister) ListSecrets(ctx context.Context) (
	secrets []coreV1.Secret, e error,
) {
	var (
		secretList *coreV1.SecretList
	)

	secretList, e = l.secrets.List(ctx,
		metaV1.ListOptions{},
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	secrets = secretList.Items

	return
}
