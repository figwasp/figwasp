package main

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/juju/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type environmentVariables struct {
	Namespace string        `env:"FIGWASP_TARGET_NAMESPACE"`
	Timeout   time.Duration `env:"FIGWASP_CLIENT_TIMEOUT"`
}

func main() {
	const (
		timeoutDefault = time.Second * 30
	)

	var (
		config  *rest.Config
		envVars environmentVariables

		swarm *FigwaspSwarm

		e error
	)

	defer func() {
		if e != nil {
			log.Fatalln(
				errors.ErrorStack(e),
			)
		}
	}()

	envVars = environmentVariables{
		Namespace: v1.NamespaceDefault,
		Timeout:   timeoutDefault,
	}

	e = env.Parse(&envVars)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	config, e = rest.InClusterConfig()
	if e != nil {
		e = errors.Trace(e)

		return
	}

	swarm, e = NewFigwaspSwarm(config,
		envVars.Namespace,
		envVars.Timeout,
	)
	if e != nil {
		e = errors.Trace(e)

		return
	}

	e = swarm.Run()
	if e != nil {
		e = errors.Trace(e)

		return
	}
}
