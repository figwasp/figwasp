package main

import (
	"context"
	"sync"
	"time"

	"github.com/juju/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"github.com/figwasp/figwasp/pkg/figwasp"
)

type FigwaspSwarm struct {
	figwasps []*Figwasp
}

func NewFigwaspSwarm(
	config *rest.Config, namespace string, timeout time.Duration,
) (
	f *FigwaspSwarm, e error,
) {
	const (
		labelSelector = "figwasp/target=true"
	)

	var (
		ctx                  context.Context
		deploymentNameLister DeploymentNameLister
		deploymentNames      []string

		credsGetter RepositoryCredentialsGetter
		restarter   RolloutRestarter
		retrievers  map[string]ImageDigestRetriever

		i int
	)

	deploymentNameLister, e = figwasp.NewLabelSelectorDeploymentNameLister(
		config,
		namespace,
		labelSelector,
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	ctx, _ = context.WithTimeout(background, timeout)

	deploymentNames, e = deploymentNameLister.ListDeploymentNames(ctx)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	credsGetter, e = newCredsGetter(config, namespace, timeout)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	restarter, e = figwasp.NewDeploymentRolloutRestarter(config, namespace)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	retrievers = make(map[string]ImageDigestRetriever)

	f = &FigwaspSwarm{
		figwasps: make([]*Figwasp,
			len(deploymentNames),
		),
	}

	for i = 0; i < len(deploymentNames); i++ {
		f.figwasps[i], e = NewFigwasp(
			config,
			namespace,
			deploymentNames[i],
			timeout,
			credsGetter,
			restarter,
			retrievers,
		)
		if e != nil {
			e = errors.Trace(e)

			return
		}
	}

	return
}

func (f *FigwaspSwarm) Run() (e error) {
	var (
		errorChannel chan error
		waitGroup    *sync.WaitGroup

		figwasp *Figwasp
	)

	errorChannel = make(chan error,
		len(f.figwasps),
	)

	waitGroup = new(sync.WaitGroup)

	waitGroup.Add(
		len(f.figwasps),
	)

	for _, figwasp = range f.figwasps {
		go figwasp.RunConcurrently(errorChannel, waitGroup)
	}

	waitGroup.Wait()

	close(errorChannel)

	for e = range errorChannel {
		if e != nil {
			e = errors.Trace(e)

			return
		}
	}

	return
}

func newCredsGetter(
	config *rest.Config, namespace string, timeout time.Duration,
) (
	credsGetter RepositoryCredentialsGetter, e error,
) {
	var (
		ctx          context.Context
		secretList   []v1.Secret
		secretLister SecretLister
	)

	secretLister, e = figwasp.NewSecretLister(config, namespace)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	ctx, _ = context.WithTimeout(background, timeout)

	secretList, e = secretLister.ListSecrets(ctx)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	credsGetter, e =
		figwasp.NewRepositoryCredentialsGetterFromKubernetesSecrets(secretList)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	return
}
