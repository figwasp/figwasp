package main

import (
	"context"

	"k8s.io/api/core/v1"

	"github.com/figwasp/figwasp/pkg/figwasp"
)

type DeploymentNameLister interface {
	ListDeploymentNames(context.Context) ([]string, error)
}

type ImageDigestRetriever interface {
	RetrieveImageDigest(string, context.Context) (string, error)
}

type ImageReferenceLister interface {
	ListImageReferences() []figwasp.ImageReference
}

type PodLister interface {
	ListPods(string, context.Context) ([]v1.Pod, error)
}

type RepositoryCredentialsGetter interface {
	GetRepositoryCredentials(string) (string, string)
}

type RolloutRestarter interface {
	RolloutRestart(string, context.Context) error
}

type SecretLister interface {
	ListSecrets(context.Context) ([]v1.Secret, error)
}
