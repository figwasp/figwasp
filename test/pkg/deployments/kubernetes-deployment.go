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
	name, kubeconfigPath string, options ...kubernetesDeploymentOption,
) (
	d *KubernetesDeployment, e error,
) {
	const (
		masterURL = ""
	)

	var (
		clientset *kubernetes.Clientset
		config    *rest.Config
		option    kubernetesDeploymentOption
	)

	config, e = clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if e != nil {
		return
	}

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		return
	}

	d = &KubernetesDeployment{
		deployments: clientset.AppsV1().Deployments(
			coreV1.NamespaceDefault,
		),

		deployment: &appsV1.Deployment{
			ObjectMeta: metaV1.ObjectMeta{
				Name:   name,
				Labels: make(map[string]string),
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
						Containers: make([]coreV1.Container, 0),
						ImagePullSecrets: make([]coreV1.LocalObjectReference,
							0,
						),
					},
				},
			},
		},

		services: clientset.CoreV1().Services(
			coreV1.NamespaceDefault,
		),

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

	for _, option = range options {
		e = option(d)
		if e != nil {
			return
		}
	}

	d.deployment, e = d.deployments.Create(
		context.Background(),
		d.deployment,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	if len(d.service.Spec.Ports) > 0 {
		d.service, e = d.services.Create(
			context.Background(),
			d.service,
			metaV1.CreateOptions{},
		)
		if e != nil {
			return
		}

	} else {
		d.service = nil
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

func (d *KubernetesDeployment) Destroy() (e error) {
	e = d.deployments.Delete(
		context.Background(),
		d.deployment.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{},
	)
	if e != nil {
		return
	}

	if d.service != nil {
		e = d.services.Delete(
			context.Background(),
			d.service.GetObjectMeta().GetName(),
			metaV1.DeleteOptions{},
		)
		if e != nil {
			return
		}
	}

	return
}

type kubernetesDeploymentOption func(*KubernetesDeployment) error

func WithReplicas(number int32) (option kubernetesDeploymentOption) {
	option = func(d *KubernetesDeployment) (e error) {
		d.deployment.Spec.Replicas = &number

		return
	}

	return
}

func WithLabel(key, value string) (option kubernetesDeploymentOption) {
	option = func(d *KubernetesDeployment) (e error) {
		d.deployment.ObjectMeta.Labels[key] = value

		d.deployment.Spec.Selector.MatchLabels[key] = value

		d.deployment.Spec.Template.ObjectMeta.Labels[key] = value

		d.service.Spec.Selector[key] = value

		return
	}

	return
}

func WithContainerWithTCPPorts(name, imageRef string, ports ...int32) (
	option kubernetesDeploymentOption,
) {
	var (
		container coreV1.Container
		i         int
	)

	container = coreV1.Container{
		Name:  name,
		Image: imageRef,
		Ports: make([]coreV1.ContainerPort,
			len(ports),
		),
	}

	for i = 0; i < len(ports); i++ {
		container.Ports[i] = coreV1.ContainerPort{
			ContainerPort: ports[i],
		}
	}

	option = func(d *KubernetesDeployment) (e error) {
		d.deployment.Spec.Template.Spec.Containers = append(
			d.deployment.Spec.Template.Spec.Containers,
			container,
		)

		for i = 0; i < len(ports); i++ {
			d.service.Spec.Ports = append(d.service.Spec.Ports,
				coreV1.ServicePort{
					Port: ports[i],
					TargetPort: intstr.FromInt(
						int(ports[i]),
					),
					NodePort: ports[i],
				},
			)
		}

		return
	}

	return
}

func WithImagePullSecrets(names ...string) (option kubernetesDeploymentOption) {
	var (
		references []coreV1.LocalObjectReference
		i          int
	)

	references = make([]coreV1.LocalObjectReference,
		len(names),
	)

	for i = 0; i < len(names); i++ {
		references[i] = coreV1.LocalObjectReference{
			Name: names[i],
		}
	}

	option = func(d *KubernetesDeployment) (e error) {
		d.deployment.Spec.Template.Spec.ImagePullSecrets = append(
			d.deployment.Spec.Template.Spec.ImagePullSecrets,
			references...,
		)

		return
	}

	return
}

func WithServiceAccount(name string) (option kubernetesDeploymentOption) {
	option = func(d *KubernetesDeployment) (e error) {
		d.deployment.Spec.Template.Spec.ServiceAccountName = name

		return
	}

	return
}

func WithHostPathVolume(name, hostPath, mountPath, containerName string,
) (
	option kubernetesDeploymentOption,
) {
	const (
		readOnly = true
	)

	var (
		container *coreV1.Container
		i         int
	)

	option = func(d *KubernetesDeployment) (e error) {
		d.deployment.Spec.Template.Spec.Volumes = append(
			d.deployment.Spec.Template.Spec.Volumes,
			coreV1.Volume{
				Name: name,
				VolumeSource: coreV1.VolumeSource{
					HostPath: &coreV1.HostPathVolumeSource{
						Path: hostPath,
					},
				},
			},
		)

		for i = 0; i < len(d.deployment.Spec.Template.Spec.Containers); i++ {
			container = &(d.deployment.Spec.Template.Spec.Containers[i])

			if container.Name != containerName {
				continue
			}

			container.VolumeMounts = append(container.VolumeMounts,
				coreV1.VolumeMount{
					Name:      name,
					ReadOnly:  readOnly,
					MountPath: mountPath,
				},
			)

			break
		}

		return
	}

	return
}
