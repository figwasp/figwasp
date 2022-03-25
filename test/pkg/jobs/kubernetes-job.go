package jobs

import (
	"context"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedBatchV1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesJob struct {
	jobs typedBatchV1.JobInterface
	job  *batchV1.Job
}

func NewKubernetesJob(
	name, kubeconfigPath string, options ...kubernetesJobOption,
) (
	j *KubernetesJob, e error,
) {
	const (
		masterURL = ""
	)

	var (
		clientset *kubernetes.Clientset
		config    *rest.Config
		option    kubernetesJobOption
	)

	config, e = clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if e != nil {
		return
	}

	clientset, e = kubernetes.NewForConfig(config)
	if e != nil {
		return
	}

	j = &KubernetesJob{
		jobs: clientset.BatchV1().Jobs(
			coreV1.NamespaceDefault,
		),

		job: &batchV1.Job{
			ObjectMeta: metaV1.ObjectMeta{
				Name: name,
			},
			Spec: batchV1.JobSpec{
				Template: coreV1.PodTemplateSpec{
					ObjectMeta: metaV1.ObjectMeta{
						Labels: make(map[string]string),
					},
					Spec: coreV1.PodSpec{
						Containers:    make([]coreV1.Container, 0),
						RestartPolicy: coreV1.RestartPolicyNever,
						ImagePullSecrets: make([]coreV1.LocalObjectReference,
							0,
						),
					},
				},
			},
		},
	}

	for _, option = range options {
		e = option(j)
		if e != nil {
			return
		}
	}

	j.job, e = j.jobs.Create(
		context.Background(),
		j.job,
		metaV1.CreateOptions{},
	)
	if e != nil {
		return
	}

	for j.job.Status.Active+j.job.Status.Succeeded+j.job.Status.Failed < 1 {
		j.job, e = j.jobs.Get(
			context.Background(),
			j.job.GetObjectMeta().GetName(),
			metaV1.GetOptions{},
		)
		if e != nil {
			return
		}
	}

	return
}

func (j *KubernetesJob) Destroy() (e error) {
	var (
		propagationPolicy metaV1.DeletionPropagation
	)

	propagationPolicy = metaV1.DeletePropagationBackground

	e = j.jobs.Delete(
		context.Background(),
		j.job.GetObjectMeta().GetName(),
		metaV1.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		},
	)
	if e != nil {
		return
	}

	return
}

type kubernetesJobOption func(*KubernetesJob) error

func WithLabel(key, value string) (option kubernetesJobOption) {
	option = func(j *KubernetesJob) (e error) {
		j.job.Spec.Template.ObjectMeta.Labels[key] = value

		return
	}

	return
}

func WithContainer(name, imageRef string, containerOptions ...containerOption) (
	option kubernetesJobOption,
) {
	const (
		envVarFormat = "%s %s"
	)

	var (
		container coreV1.Container
		i         int
	)

	container = coreV1.Container{
		Name:  name,
		Image: imageRef,
	}

	for i = 0; i < len(containerOptions); i++ {
		containerOptions[i](&container)
	}

	option = func(j *KubernetesJob) (e error) {
		j.job.Spec.Template.Spec.Containers = append(
			j.job.Spec.Template.Spec.Containers,
			container,
		)

		return
	}

	return
}

func WithImagePullSecrets(names ...string) (option kubernetesJobOption) {
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

	option = func(j *KubernetesJob) (e error) {
		j.job.Spec.Template.Spec.ImagePullSecrets = append(
			j.job.Spec.Template.Spec.ImagePullSecrets,
			references...,
		)

		return
	}

	return
}

func WithServiceAccount(name string) (option kubernetesJobOption) {
	option = func(j *KubernetesJob) (e error) {
		j.job.Spec.Template.Spec.ServiceAccountName = name

		return
	}

	return
}

func WithHostPathVolume(name, hostPath string) (option kubernetesJobOption) {
	option = func(j *KubernetesJob) (e error) {
		j.job.Spec.Template.Spec.Volumes = append(
			j.job.Spec.Template.Spec.Volumes,
			coreV1.Volume{
				Name: name,
				VolumeSource: coreV1.VolumeSource{
					HostPath: &coreV1.HostPathVolumeSource{
						Path: hostPath,
					},
				},
			},
		)

		return
	}

	return
}

type containerOption func(*coreV1.Container)

func WithEnvironmentVariable(key, value string) (option containerOption) {
	option = func(c *coreV1.Container) {
		c.Env = append(c.Env,
			coreV1.EnvVar{
				Name:  key,
				Value: value,
			},
		)

		return
	}

	return
}

func WithVolumeMount(volumeName, mountPath string) (option containerOption) {
	const (
		readOnly = true
	)

	option = func(c *coreV1.Container) {
		c.VolumeMounts = append(c.VolumeMounts,
			coreV1.VolumeMount{
				Name:      volumeName,
				ReadOnly:  readOnly,
				MountPath: mountPath,
			},
		)

		return
	}

	return
}
