Feature: Alduin

    As a DevOps practitioner deploying containers to a Kubernetes cluster,
    I want a sidecar that triggers a rolling restart of the deployment
    whenever a later version of a currently-deployed image is available
    so that software updates are automatically applied with minimal setup.

    Background:
        Given there is a Kubernetes Deployment
        And the Deployment has a PodTemplateSpec describing a Pod
        And the Pod runs one or more Linux containers from container images
        And a container image is pulled from a container image repository
        And a container image is uniquely identified by an image digest
        And a container image is normally referred to by a name and tag
        And all versions of a container image share the same name
        And a later version of a container image inherits a predecessor's tag
        And the image digest indexed by a name-tag pair changes with updates

    Scenario:
        Given a Deployment running some container(s) including Alduin
        When I push an image with the same name and tag as one of the containers
        And the image has a digest not equal to the one deployed
        Then Alduin would trigger a rolling restart of the Deployment
        And I would see Pods running containers from the new image created
        And I would see Pods running containers from the old image deleted
