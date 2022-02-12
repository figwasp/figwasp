```gherkin
Feature: Alduin

    As a Kubernetes administrator deploying container applications to a cluster,
    I want a rolling restart of a deployment to be automatically triggered
    whenever the tag of a currently-deployed image is inherited by another image
    so that the deployment is always up-to-date without manual intervention.

    Scenario:
        Given there is a container image repository
        And in the repository there is an image of a container
        And the container is a HTTP server with a GET endpoint
        And the endpoint responds to requests with a fixed HTTP status code
        And the status code is preset via a container image build argument
        And there is a Kubernetes cluster
        And the server is deployed to the cluster using a Kubernetes deployment
        And the image of the server is pulled from the repository
        And the endpoint is exposed using a Kubernetes service and ingress
        And Alduin is running in the cluster
        And Alduin is authenticated as a Kubernetes service account
        And the service account is authorised to get and patch deployments

        When I rebuild the image so that it returns a different status code
        And I transfer to the new image the tag of the existing image
        And I push the new image to the repository
        And I allow time for a rolling restart of the deployment to complete
        And I send a request to the endpoint

        Then I should see in the response to the request the new status code
```
