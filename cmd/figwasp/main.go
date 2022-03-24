package main

import (
	"context"
	"log"

	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"github.com/figwasp/figwasp/pkg/deployments"
	"github.com/figwasp/figwasp/pkg/images"
	"github.com/figwasp/figwasp/pkg/pods"
	"github.com/figwasp/figwasp/pkg/repositories"
	"github.com/figwasp/figwasp/pkg/secrets"
)

func main() {
	const (
		deploymentName = "http-status-code-server" //TODO
	)

	var (
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

	ctx = context.Background() //TODO

	config, e = rest.InClusterConfig()
	if e != nil {
		log.Fatalln(e)
	}

	restarter, e = deployments.NewDeploymentRolloutRestarter(
		config,
		v1.NamespaceDefault,
	)
	if e != nil {
		log.Fatalln(e)
	}

	podLister, e = pods.NewDeploymentPodLister(config, v1.NamespaceDefault)
	if e != nil {
		log.Fatalln(e)
	}

	podList, e = podLister.ListPods(deploymentName, ctx)
	if e != nil {
		log.Fatalln(e)
	}

	refLister, e = images.NewImageReferenceListerFromPods(podList)
	if e != nil {
		log.Fatalln(e)
	}

	refList = refLister.ListImageReferences()

	secretLister, e = secrets.NewSecretLister(config, v1.NamespaceDefault)
	if e != nil {
		log.Fatalln(e)
	}

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

		digest, e = retriever.RetrieveImageDigest(
			reference.NamedAndTagged(),
			ctx,
		)
		if e != nil {
			log.Fatalln(e)
		}

		if digest != reference.ImageDigest() {
			e = restarter.RolloutRestart(deploymentName, ctx)
			if e != nil {
				log.Fatalln(e)
			}

			return
		}
	}
}
