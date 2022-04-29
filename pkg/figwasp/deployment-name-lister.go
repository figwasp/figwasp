package figwasp

import (
	"context"

	"github.com/juju/errors"
	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedAppsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
)

type labelSelectorDeploymentNameLister struct {
	deployments   typedAppsV1.DeploymentInterface
	labelSelector string
}

func NewLabelSelectorDeploymentNameLister(
	config *rest.Config, namespace, labelSelector string,
) (
	l *labelSelectorDeploymentNameLister, e error,
) {
	var (
		clientset *kubernetes.Clientset
	)

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	l = &labelSelectorDeploymentNameLister{
		deployments:   clientset.AppsV1().Deployments(namespace),
		labelSelector: labelSelector,
	}

	return
}

func (l *labelSelectorDeploymentNameLister) ListDeploymentNames(
	ctx context.Context,
) (
	deploymentNames []string, e error,
) {
	var (
		deploymentList *appsV1.DeploymentList

		i int
	)

	deploymentList, e = l.deployments.List(ctx,
		metaV1.ListOptions{
			LabelSelector: l.labelSelector,
		},
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	deploymentNames = make([]string,
		len(deploymentList.Items),
	)

	for i = 0; i < len(deploymentList.Items); i++ {
		deploymentNames[i] = deploymentList.Items[i].GetObjectMeta().GetName()
	}

	return
}
