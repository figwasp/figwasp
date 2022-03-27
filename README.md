# Keep Kubernetes Deployments up-to-date with the `:latest`* container images
Practising continuous integration in a highly-automated environment means that
new versions of multiple build artifacts are generated with every incremental
change to the codebase. In the context of cloud-native microservices, these
build artifacts often take the form of Linux container images that are deployed
and managed via an orchestrator such as Kubernetes. Resources would be
provisioned for a test cluster, to which artifacts built from the latest commit
pushed to the trunk of the code repository would be deployed to undergo
acceptance tests. Artifacts that pass the test would be tagged as releasable
for deployment to a production cluster.

While Kubernetes can be configured to pull the latest image with a given tag
when a Deployment is restarted, it does not provide a means of watching
container image repositories for changes in tags. A restart had to be triggered
by an administrator using the command-line interface, or by build automation
calling the Kubernetes API. The former is a manual bottleneck in an otherwise
automated workflow, while the latter necessitates the delegation of cluster
administrator privileges and credentials. Figwasp offers an elegant alternative
to the existing "push" model and can easily be employed as a Kubernetes CronJob
with minimal configuration.

## User Story
    As a Kubernetes administrator deploying container applications to a cluster,
    I want a rolling restart of a deployment to be automatically triggered
    whenever the tag of a currently-deployed image is inherited by another image
    so that the deployment is always up-to-date without manual intervention.

### [Kubernetes issue #33664](https://github.com/kubernetes/kubernetes/issues/33664) "Force pods to re-pull an image without changing the image tag"
@yissachar:
> Problem
>
> A frequent question that comes up on Slack and Stack Overflow is how to
> trigger an update to a Deployment/RS/RC when the image tag hasn't changed but
> the underlying image has.
>
> Consider:
>
> 1. There is an existing Deployment with image `foo:latest`
> 1. User builds a new image `foo:latest`
> 1. User pushes `foo:latest` to their registry
> 1. User wants to do something here to tell the Deployment to pull the new
> image and do a rolling-update of existing pods
>
> The problem is that there is no existing Kubernetes mechanism which properly
> covers this.

### [StackOverflow question](https://stackoverflow.com/questions/45905999) "Kubernetes: --image-pull-policy always does not work"
DenCowboy:
> I have a Kubernetes deployment which uses image: `test:latest` (not real
> image name but it's the latest tag). This image is on docker hub. I have just
> pushed a new version of `test:latest` to dockerhub. I was expecting a new
> deployment of my pod in Kubernetes but nothing happends.
>
> I've created my deployment like this:
>
> ```
> kubectl run sample-app --image=`test:latest` --namespace=sample-app --image-pull-policy Always
> ```
>
> Why isn't there a new deployment triggered after the push of a new image?

### [StackOverflow question](https://stackoverflow.com/questions/65277807) "How to automatically restart pods when a new image ready"
user1739211:
> I was expecting that every time I would push a new image with the tag latest,
> the pods would be automatically killed and restart using the new images.
>
> I tried the rollout command
>
> ```
> kubectl rollout restart deploy simpleapp-direct
> ```
>
> The pods restart as I wanted.
>
> However, I don't want to run this command every time there is a new latest
> build. How can I handle this situation ?.

### [StackOverflow question](https://stackoverflow.com/questions/40366192) "Kubernetes how to make Deployment to update image"
Andriy Kopachevskyy:
> I do have deployment with single pod, with my custom docker image like:
>
> ```yaml
> containers:
>   - name: mycontainer
>     image: myimage:latest
> ```
>
> During development I want to push new latest version and make Deployment
> updated.

### [StackOverflow question](https://stackoverflow.com/questions/41735829) "Update a Deployment image in Kubernetes"
Yuval Simhon:
> I saw this in the documentation:
>
> > Note: a Deployment’s rollout is triggered if and only if the Deployment’s
> > pod template (i.e. .spec.template) is changed.
>
>
> I'm searching for an easy way/workaround to automate the flow: Build
> triggered > a new Docker image is pushed (withoud version changing) >
> Deployment will update the pod > service will expose the new pod.

## Behaviour-Driven Specifications
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
