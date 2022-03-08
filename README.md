```gherkin
Feature: Alduin

    As a Kubernetes administrator deploying container applications to a cluster,
    I want a rolling restart of a deployment to be automatically triggered
    whenever the tag of a currently-deployed image is inherited by another image
    so that the deployment is always up-to-date without manual intervention.

    Scenario:
        Given there is a container image repository
        And in the repository there is an image of a server
        And there is a Kubernetes cluster
        And the server is deployed to the cluster by means of a Deployment
        And Alduin is running in the cluster with the relevant permissions

        When I rebuild the server image so that it behaves differently
        And I push the new image to the repository with the same tag as the old
        And I allow time for a rolling restart of the deployment to complete
        And I send a request to the server and receive a response

        Then I should observe in the response evidence of the new behaviour
```
