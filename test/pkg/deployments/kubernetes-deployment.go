package deployments

import (
	"context"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	typedAppsV1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	typedCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesDeployment struct {
	deployments typedAppsV1.DeploymentInterface
	deployment  *appsV1.Deployment
	services    typedCoreV1.ServiceInterface
	service     *coreV1.Service
}

func NewKubernetesDeployment(
	name, kubeConfigPath string,
) (
	d *KubernetesDeployment, e error,
) {
	const (
		hostNetwork = true
		masterURL   = ""
		replicas    = 1
	)

	var (
		clientset *kubernetes.Clientset
		config    *rest.Config
	)

	config, e = clientcmd.BuildConfigFromFlags(masterURL, kubeConfigPath)
	if e != nil {
		return
	}

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		return
	}

	d = &KubernetesDeployment{
		deployments: clientset.AppsV1().Deployments(coreV1.NamespaceDefault),
		deployment: &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Spec: appsV1.DeploymentSpec{
				Selector: &metaV1.LabelSelector{
					MatchLabels: make(map[string]string),
				},
				Template: coreV1.PodTemplateSpec{
					ObjectMeta: metaV1.ObjectMeta{
						Labels: make(map[string]string),
					},
					Spec: coreV1.PodSpec{
						Containers:  make([]coreV1.Container, 0),
						HostNetwork: hostNetwork,
					},
				},
			},
		},
		services: clientset.CoreV1().Services(coreV1.NamespaceDefault),
		service: &coreV1.Service{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Spec: coreV1.ServiceSpec{
				Ports:    make([]coreV1.ServicePort, 0),
				Selector: make(map[string]string),
				Type:     coreV1.ServiceTypeNodePort,
			},
		},
	}

	d.setReplicas(replicas)

	return
}

func (d *KubernetesDeployment) setReplicas(replicas int32) {
	d.deployment.Spec.Replicas = &replicas

	return
}

func (d *KubernetesDeployment) SetLabel(key, value string) {
	d.deployment.Spec.Selector.MatchLabels[key] = value

	d.deployment.Spec.Template.ObjectMeta.Labels[key] = value

	d.service.Spec.Selector[key] = value

	return
}

func (d *KubernetesDeployment) AddSingleTCPPortContainer(
	name, imageRef string, port int32,
) {
	d.deployment.Spec.Template.Spec.Containers = append(
		d.deployment.Spec.Template.Spec.Containers,
		coreV1.Container{
			Name:  name,
			Image: imageRef,
			Ports: []coreV1.ContainerPort{
				{
					HostPort:      port,
					ContainerPort: port,
				},
			},
		},
	)

	d.service.Spec.Ports = append(d.service.Spec.Ports,
		coreV1.ServicePort{
			Port: port,
			TargetPort: intstr.FromInt(
				int(port),
			),
		},
	)

	return
}

func (d *KubernetesDeployment) Create() (e error) {
	d.deployment, e = d.deployments.Create(
		context.Background(),
		d.deployment,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	d.service, e = d.services.Create(
		context.Background(),
		d.service,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	for d.deployment.Status.AvailableReplicas == 0 {
		d.deployment, e = d.deployments.Get(
			context.Background(),
			d.deployment.GetObjectMeta().GetName(),
			metaV1.GetOptions{},
		)
		if e != nil {
			return
		}
	}

	return
}

func (d *KubernetesDeployment) IPAddress() (address string, e error) {
	for d.service.Spec.ClusterIP == "" {
		d.service, e = d.services.Get(
			context.Background(),
			d.service.GetObjectMeta().GetName(),
			metaV1.GetOptions{},
		)
		if e != nil {
			return
		}
	}

	address = d.service.Spec.ClusterIP

	return
}

func (d *KubernetesDeployment) Delete() (e error) {
	e = d.deployments.Delete(
		context.Background(),
		d.deployment.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{},
	)
	if e != nil {
		return
	}

	e = d.services.Delete(
		context.Background(),
		d.service.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{},
	)
	if e != nil {
		return
	}

	return
}
