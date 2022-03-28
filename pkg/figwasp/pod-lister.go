package figwasp

import (
	"context"

	"github.com/juju/errors"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedAppsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type deploymentPodLister struct {
	deployments typedAppsV1.DeploymentInterface
	replicaSets typedAppsV1.ReplicaSetInterface
	pods        typedCoreV1.PodInterface
}

func NewDeploymentPodLister(config *rest.Config, namespace string) (
	l *deploymentPodLister, e error,
) {
	var (
		clientset *kubernetes.Clientset
	)

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	l = &deploymentPodLister{
		deployments: clientset.AppsV1().Deployments(namespace),
		replicaSets: clientset.AppsV1().ReplicaSets(namespace),
		pods:        clientset.CoreV1().Pods(namespace),
	}

	return
}

func (l *deploymentPodLister) ListPods(
	deploymentName string, ctx context.Context,
) (
	pods []coreV1.Pod, e error,
) {
	var (
		deployment      *appsV1.Deployment
		pod             coreV1.Pod
		podList         *coreV1.PodList
		replicaSet      appsV1.ReplicaSet
		replicaSetFound bool
		replicaSetList  *appsV1.ReplicaSetList
	)

	deployment, e = l.deployments.Get(ctx,
		deploymentName,
		metaV1.GetOptions{},
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	replicaSetList, e = l.replicaSets.List(ctx,
		metaV1.ListOptions{},
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	for _, replicaSet = range replicaSetList.Items {
		if metaV1.IsControlledBy(&replicaSet, deployment) {
			replicaSetFound = true

			break
		}
	}

	if !replicaSetFound {
		return
	}

	podList, e = l.pods.List(ctx,
		metaV1.ListOptions{},
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	for _, pod = range podList.Items {
		if metaV1.IsControlledBy(&pod, &replicaSet) {
			pods = append(pods, pod)
		}
	}

	return
}
