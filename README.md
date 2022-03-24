```gherkin
Feature: Figwasp

    As a Kubernetes administrator deploying container applications to a cluster,
    I want a rolling restart of a deployment to be automatically triggered
    whenever the tag of a currently-deployed image is inherited by another image
    so that the deployment is always up-to-date without manual intervention.

    Scenario:
        Given there is a container image repository
        And in the repository there is a container image of a server
        And there is a Kubernetes cluster
        And in the cluster there is a Kubernetes Deployment of that server
        And there is a client to the server obtaining its response to a request

        When there is a new container image of the server in the repository
        And Figwasp is run as a Kubernetes Job or CronJob in the cluster

        Then the client should detect a corresponding change in server response
```
