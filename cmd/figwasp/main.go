package main

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type environmentVariables struct {
	Namespace  string        `env:"FIGWASP_TARGET_NAMESPACE"`
	Deployment string        `env:"FIGWASP_TARGET_DEPLOYMENT,notEmpty"`
	Timeout    time.Duration `env:"FIGWASP_API_CLIENT_TIMEOUT"`
}

func main() {
	const (
		timeoutDefault = time.Second * 30
	)

	var (
		config  *rest.Config
		envVars environmentVariables

		figwasp *Figwasp

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

	config, e = rest.InClusterConfig()
	if e != nil {
		log.Fatalln(e)
	}

	figwasp, e = NewFigwasp(config,
		envVars.Namespace,
		envVars.Deployment,
		envVars.Timeout,
	)
	if e != nil {
		log.Fatalln(e)
	}

	e = figwasp.Run()
	if e != nil {
		log.Fatalln(e)
	}
}
