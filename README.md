```gherkin
Feature: Alduin

    As a Kubernetes administrator deploying container applications to a cluster,
    I want a rolling restart of a deployment to be automatically triggered
    whenever the tag of a currently-deployed image is inherited by another image
    so that the deployment is always up-to-date without manual intervention.

    Scenario:
        Given there is a Kubernetes cluster
        And Alduin is running in any pod in that cluster
        And Alduin is authenticated as a Kubernetes service account
        And that service account is authorised to get and patch deployments
        And a Kubernetes deployment is created in that cluster
        And the deployment has a PodTemplateSpec describing a pod
        And the pod runs a container application serving a message over HTTP
        And the application is exposed through a Kubernetes service and ingress

        When I push an image of that application configured with a new message
        And the new image inherits the name and tag of the existing image
        And I allow time for a rolling restart of the deployment to complete
        And I access the exposed application by sending to it a HTTP request

        Then I should see the new message in the HTTP response returned
```
