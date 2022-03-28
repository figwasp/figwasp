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

var (
	background context.Context = context.Background()
)

type Figwasp struct {
	credsGetter RepositoryCredentialsGetter
	references  []figwasp.ImageReference
	restarter   RolloutRestarter
	retrievers  map[string]ImageDigestRetriever

	deployment string
	timeout    time.Duration
}

func NewFigwasp(
	config *rest.Config, namespace, deployment string, timeout time.Duration,
) (
	f *Figwasp, e error,
) {
	var (
		reference figwasp.ImageReference
		refLister ImageReferenceLister
	)

	refLister, e = newRefLister(config, namespace, deployment, timeout)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	f = &Figwasp{
		references: refLister.ListImageReferences(),
		retrievers: make(map[string]ImageDigestRetriever),

		deployment: deployment,
		timeout:    timeout,
	}

	f.credsGetter, e = newCredsGetter(config, namespace, timeout)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	for _, reference = range f.references {
		e = f.addRetriever(reference.RepositoryAddress)
		if e != nil {
			e = errors.Trace(e)

			return
		}
	}

	f.restarter, e = figwasp.NewDeploymentRolloutRestarter(config, namespace)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	return
}

func (f *Figwasp) Run() (e error) {
	var (
		reference figwasp.ImageReference

		waitGroup *sync.WaitGroup

		failure chan error
		nothing chan struct{}
		trigger chan struct{}
	)

	failure = make(chan error)
	nothing = make(chan struct{})
	trigger = make(chan struct{})

	waitGroup = new(sync.WaitGroup)

	waitGroup.Add(
		len(f.references),
	)

	for _, reference = range f.references {
		go f.retrieveAndCompareImageDigest(reference,
			waitGroup,
			trigger,
			failure,
		)
	}

	go wait(waitGroup, nothing)

	select {
	case <-trigger:
		f.rolloutRestart()

	case e = <-failure:
		return

	case <-nothing:
		return
	}

	return
}

func (f *Figwasp) addRetriever(repositoryAddress string) (e error) {
	var (
		found     bool
		retriever ImageDigestRetriever
	)

	_, found = f.retrievers[repositoryAddress]
	if found {
		return
	}

	retriever, e = figwasp.NewImageDigestRetriever(
		figwasp.WithBasicAuthentication(
			f.credsGetter.GetRepositoryCredentials(repositoryAddress),
		),
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	f.retrievers[repositoryAddress] = retriever

	return
}

func (f *Figwasp) retrieveAndCompareImageDigest(
	reference figwasp.ImageReference, waitGroup *sync.WaitGroup,
	trigger chan<- struct{}, failure chan<- error,
) {
	var (
		ctx    context.Context
		digest string
		e      error
	)

	ctx, _ = context.WithTimeout(background, f.timeout)

	digest, e = f.retrievers[reference.RepositoryAddress].RetrieveImageDigest(
		reference.NamedAndTagged,
		ctx,
	)
	if e != nil {
		failure <- errors.Trace(e)

		return
	}

	if digest != reference.ImageDigest {
		close(trigger)

		return
	}

	waitGroup.Done()

	return
}

func (f *Figwasp) rolloutRestart() (e error) {
	var (
		ctx context.Context
	)

	ctx, _ = context.WithTimeout(background, f.timeout)

	e = f.restarter.RolloutRestart(f.deployment, ctx)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	return
}

func newRefLister(
	config *rest.Config, namespace, deployment string, timeout time.Duration,
) (
	refLister ImageReferenceLister, e error,
) {
	var (
		ctx       context.Context
		podList   []v1.Pod
		podLister PodLister
	)

	podLister, e = figwasp.NewDeploymentPodLister(config, namespace)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	ctx, _ = context.WithTimeout(background, timeout)

	podList, e = podLister.ListPods(deployment, ctx)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	refLister, e = figwasp.NewImageReferenceListerFromPods(podList)
	if e != nil {
		e = errors.Trace(e)

		return
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

func wait(waitGroup *sync.WaitGroup, nothing chan<- struct{}) {
	waitGroup.Wait()

	close(nothing)

	return
}
