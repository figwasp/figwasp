package main

import (
	"context"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"github.com/figwasp/figwasp/pkg/deployments"
	"github.com/figwasp/figwasp/pkg/images"
	"github.com/figwasp/figwasp/pkg/pods"
	"github.com/figwasp/figwasp/pkg/repositories"
	"github.com/figwasp/figwasp/pkg/secrets"
)

type environmentVariables struct {
	Deployment string        `env:"FIGWASP_TARGET_DEPLOYMENT,notEmpty"`
	Namespace  string        `env:"FIGWASP_TARGET_NAMESPACE"`
	Timeout    time.Duration `env:"FIGWASP_API_CLIENT_TIMEOUT"`
}

func main() {
	const (
		timeoutDefault = time.Second * 30
	)

	var (
		background context.Context
		envVars    environmentVariables

		config *rest.Config
		ctx    context.Context

		credsGetter  repositories.RepositoryCredentialsGetter
		podLister    pods.PodLister
		refLister    images.ImageReferenceLister
		restarter    deployments.RolloutRestarter
		secretLister secrets.SecretLister

		podList    []v1.Pod
		refList    []images.ImageReference
		retrievers map[string]images.ImageDigestRetriever
		secretList []v1.Secret

		reference images.ImageReference
		retriever images.ImageDigestRetriever

		digest string
		found  bool

		e error
	)

	envVars = environmentVariables{
		Namespace: v1.NamespaceDefault,
		Timeout:   timeoutDefault,
	}

	e = env.Parse(&envVars)
	if e != nil {
		log.Fatalln(e)
	}

	background = context.Background()

	config, e = rest.InClusterConfig()
	if e != nil {
		log.Fatalln(e)
	}

	restarter, e = deployments.NewDeploymentRolloutRestarter(
		config,
		envVars.Namespace,
	)
	if e != nil {
		log.Fatalln(e)
	}

	podLister, e = pods.NewDeploymentPodLister(config, envVars.Namespace)
	if e != nil {
		log.Fatalln(e)
	}

	ctx, _ = context.WithTimeout(background, envVars.Timeout)

	podList, e = podLister.ListPods(envVars.Deployment, ctx)
	if e != nil {
		log.Fatalln(e)
	}

	refLister, e = images.NewImageReferenceListerFromPods(podList)
	if e != nil {
		log.Fatalln(e)
	}

	refList = refLister.ListImageReferences()

	secretLister, e = secrets.NewSecretLister(config, envVars.Namespace)
	if e != nil {
		log.Fatalln(e)
	}

	ctx, _ = context.WithTimeout(background, envVars.Timeout)

	secretList, e = secretLister.ListSecrets(ctx)
	if e != nil {
		log.Fatalln(e)
	}

	credsGetter, e =
		repositories.NewRepositoryCredentialsGetterFromKubernetesSecrets(
			secretList,
		)
	if e != nil {
		log.Fatalln(e)
	}

	retrievers = make(map[string]images.ImageDigestRetriever)

	for _, reference = range refList {
		_, found = retrievers[reference.RepositoryAddress()]
		if found {
			continue
		}

		retriever, e = images.NewImageDigestRetriever(
			images.WithBasicAuthentication(
				credsGetter.GetRepositoryCredentials(
					reference.RepositoryAddress(),
				),
			),
		)
		if e != nil {
			log.Fatalln(e)
		}

		retrievers[reference.RepositoryAddress()] = retriever
	}

	for _, reference = range refList {
		retriever = retrievers[reference.RepositoryAddress()]

		ctx, _ = context.WithTimeout(background, envVars.Timeout)

		digest, e = retriever.RetrieveImageDigest(
			reference.NamedAndTagged(),
			ctx,
		)
		if e != nil {
			log.Fatalln(e)
		}

		if digest != reference.ImageDigest() {
			ctx, _ = context.WithTimeout(background, envVars.Timeout)

			e = restarter.RolloutRestart(envVars.Deployment, ctx)
			if e != nil {
				log.Fatalln(e)
			}

			return
		}
	}
}
