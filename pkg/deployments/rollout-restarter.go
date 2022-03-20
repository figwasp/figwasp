package deployments

import (
	"context"
	"time"

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedAppsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
)

type RolloutRestarter interface {
	RolloutRestart(string, context.Context) error
}

type deploymentRolloutRestarter struct {
	deployments typedAppsV1.DeploymentInterface
}

func NewDeploymentRolloutRestarter(config *rest.Config, namespace string) (
	r *deploymentRolloutRestarter, e error,
) {
	var (
		clientset *kubernetes.Clientset
	)

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		return
	}

	r = &deploymentRolloutRestarter{
		deployments: clientset.AppsV1().Deployments(namespace),
	}

	return
}

func (r *deploymentRolloutRestarter) RolloutRestart(
	deploymentName string, ctx context.Context,
) (
	e error,
) {
	const (
		annotationKey = "alduin/restartedAt"
	)

	var (
		deployment *appsV1.Deployment
	)

	deployment, e = r.deployments.Get(ctx,
		deploymentName,
		metaV1.GetOptions{},
	)
	if e != nil {
		return
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations =
			make(map[string]string)
	}

	deployment.Spec.Template.ObjectMeta.Annotations[annotationKey] =
		time.Now().Format(time.RFC3339)

	deployment, e = r.deployments.Update(ctx,
		deployment,
		metaV1.UpdateOptions{},
	)
	if e != nil {
		return
	}

	return
}
