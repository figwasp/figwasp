package pods

import (
	"context"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedAppsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type PodLister interface {
	ListPodsControlledByDeployment(string, context.Context) (
		[]coreV1.Pod, error,
	)
}

type podLister struct {
	deployments typedAppsV1.DeploymentInterface
	replicaSets typedAppsV1.ReplicaSetInterface
	pods        typedCoreV1.PodInterface
}

func NewPodLister(config *rest.Config, namespace string) (
	l *podLister, e error,
) {
	var (
		clientset *kubernetes.Clientset
	)

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		return
	}

	l = &podLister{
		deployments: clientset.AppsV1().Deployments(namespace),
		replicaSets: clientset.AppsV1().ReplicaSets(namespace),
		pods:        clientset.CoreV1().Pods(namespace),
	}

	return
}

func (l *podLister) ListPodsControlledByDeployment(
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
		return
	}

	replicaSetList, e = l.replicaSets.List(ctx,
		metaV1.ListOptions{},
	)
	if e != nil {
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
		return
	}

	for _, pod = range podList.Items {
		if metaV1.IsControlledBy(&pod, &replicaSet) {
			pods = append(pods, pod)
		}
	}

	return
}
