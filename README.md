`alduin` is a Kubernetes sidecar that automatically updates a deployment
when a newer image inherits the name and tag of a currently-deployed
image.

Behaviour-Driven Specifications
===============================
```gherkin
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
```

Design
======

Identify Target Deployment(s)
-----------------------------

To fulfill its purpose of updating a Kubernetes *deployment*, `alduin`
must obtain access to its target. A deployment is most straightforwardly
addressed by its name, and the Kubernetes *namespace* in which the
deployment could be found. To obtain these values, a
`DeploymentIdentifier` interface would be defined to provide a method
`IdentifyDeployment`. It would be implemented by an
`EnvironmentVariableDeploymentIdentifier` that reads the namespace and
deployment names from environment variables assigned with those values.

In the event that a deployment is not specified, `alduin` would target
all deployments in a namespace, which if also left undefined must be
assumed to be the `default` namespace. For that purpose a
`DeploymentLister` interface would be defined to provide a method
`ListDeployments`, which takes a namespace as an argument and returns a
collection of `Deployment` interfaces representing deployments in that
namespace. `DeploymentLister` would be implemented by a
`V1APIClientDeploymentLister` by means of a Go client to Version 1 of
the Kubernetes API.

`DeploymentLister` would also provide a second method `GetDeployment`,
which takes a namespace and deployment name as arguments and returns a
collection containing only one `Deployment` corresponding to that name.

### `V1APIClientDeploymentLister`

A Go package `github.com/kubernetes/client-go/kubernetes` (hereafter
abbrieviated as "`...kubernetes`") provides a function `NewForConfig`
that creates a new `Clientset` for the given config[^2]. `Clientset` has
a method `AppsV1` that returns an `AppsV1Interface` defined in package
`...kubernetes/typed/apps/v1`. As a composite interface embedding
`DeploymentsGetter`, `AppsV1Interface` implements the method
`Deployments`, which returns a `DeploymentInterface`. The interface has
methods `List` and `Get`, which `V1APIClientDeploymentLister` would
invoke to obtain return values of type `DeploymentList` and `Deployment`
respectively. The former is a collection of the latter, as their names
suggest.

`DeploymentInterface` is a reusable resource that should be initialised
and then passed to the constructor of `V1APIClientDeploymentLister` to
be stored in an unexported field of the struct. It can foreseeably be
used by other interfaces in `alduin` to make other deployment-related
calls to the Kubernetes API.

### The `Deployment` Interface

The methods of `alduin` interfaces should not interact directly with
structs returned by methods of `DeploymentInterface` to avoid tight
coupling with a particular client implementation (and version) of the
Kubernetes API. Instead, a type `V1Deployment` wrapping the `Deployment`
struct would be declared. It would carry the methods that satisfy the
custom `Deployment` interface decoupling the rest of `alduin` from
Version 1 of the Kubernetes Go client.

The `Deployment` interface would also provide a method `Name` to allow
convenient access to deployment names.

Identify Images to Monitor
--------------------------

An `ImageLister` interface would be defined to provide a method
`ListImages`, which returns a collection of container image identifiers.
These identifiers are strings as is the case in `containerd`[^3]. They
contain an image name and optionally an image repository hostname, port
number, path, as well as a tag, following the convention popularised by
Docker and adopted by Kubernetes.

Conceptually, there are many ways to obtain such a list of image
identifiers, including the extreme case of reading the list from an
environment variable, file or other forms of input. However, since a
deployment *spec* contains information about images, it is efficient to
extract the list of deployed images from that. The `Deployment`
interface would therefore embed the `ImageLister` interface.

### `Deployment.ListImages`

Listing the images deployed in a deployment is merely a matter of
accessing field `.Spec.Template.Spec.Containers` of a `Deployment`
struct, a slice of `Container` structs defined in package
`github.com/kubernetes/api/v1`. The target string `Container.Image`
would be appended to the list of image identifiers that `ListImages`
would return.

Periodically Retrieve Image Digests
-----------------------------------

The Open Container Initiative (OCI) image specifications state that a
digest is a string that uniquely identifies image contents.

An `ImageDigestRetriever` interface would be defined to provide a method
`RetrieveImageDigest`, which returns the digest of a container image
given its identifier. Although some implementations may require
credentials for authenticating to image repositories, the interface does
not assume the necessity nor form of such additional arguments.

A `DockerClientImageDigestRetriever` would implement the interface
utilising the official Go client for the Docker Engine API.

### `DockerClientImageDigestRetriever`

The Docker Engine API defines an endpoint `distribution/<name>/json`,
which returns among other things the digest of an image given its name.
Package `github.com/moby/moby/client` provides a struct `Client` that
has a method `DistributionInspect` for accessing that endpoint. It
expects a `base64url`-encoded JSON "identity token" of format
`{"identitytoken": "<token>"}`, where `<token>` is a value that can be
obtained by making a `POST` request to the `auth` endpoint using another
`Client` method `RegistryLogin`.

Maintain Register of Image Digests
----------------------------------

An `ImageDigestRegister` interface would be defined to provide a method
`RegisterImageDigest` that takes an image identifier and an image digest
as arguments, and associates the digest with the identifier.

`MapImageDigestRegister`, a Go map of type `map[string]string`, would
implement the interface.

Compare Present and Historical Image Digests
--------------------------------------------

Whereas the OCI image specifications define a digest format and allow
several digest algorithms, `alduin` is only concerned with the absolute
value of the digests, since an image would always be pulled from the
same repository across versions, and the digest format and algorithm is
unlikely to vary.

`ImageDigestRegister` would provide a method `CompareAgainstRecords`
that takes an image identifier and an image digest as arguments, and
returns a boolean indicating whether the digest differs from that in
history.

Update Deployment(s)
--------------------

A `DeploymentUpdater` interface would be defined to provide
`UpdateDeployment`, a method that takes the name of a deployment as an
argument and causes Kubernetes to perform a rolling restart of that
deployment, downloading the latest image for each container in it if
that image has changed, contingent on the `imagePullPolicy` of those
containers set to `Always`.

A `PatchingDeploymentUpdater` inspired by the workings behind the
Kubernetes command-line tool `kubectl` (discussed in greater detail
below) would implement the `DeploymentUpdater` interface.

### `PatchingDeploymentUpdater`

`kubectl` accepts a command `rollout restart deployment/<name>`[^4] that
triggers a rolling restart of a named deployment. Under the hood,
`kubectl` sends `PATCH` requests to the Kubernetes (HTTP) API that
trigger a rollout of the deployment by changing its pod template.
Specifically, it updates the annotation
`kubectl.kubernetes.io/restartedAt` in the template metadata to an
RFC3339 string representing the current time. Evidence of such behaviour
is found in the unexported `defaultObjectRestarter` function in the
package `polymorphichelpers` in the `kubectl` source. The function is
passed as an argument to another that prepares the patch request, itself
called by method `RunRestart` from package `cmd/rollout`, the method
behind `rollout restart`.

`PatchingDeploymentUpdater` could use the `Patch` method supplied by a
`DeploymentInterface` to send a `PATCH` request with a JSON body similar
to the following, essentially mimicking the behaviour of `kubectl`.

```json
    {
       "spec": {
          "template": {
             "metadata": {
                "annotations": {
                   "alduin/restartedAt": "<timestamp>"
                }
             }
          }
       }
    }
```

Usage
-----

The user would need to create a Role or ClusterRole with permissions to
`get` and `patch` deployments. The role would be bound to a service
account, and the `psychocomp` sidecar would be authenticated as that
service account.

Testing
=======

Mock Kubernetes Cluster
-----------------------

`kind`

[^1]: Or any other tag, for that matter.

[^2]: The package `github.com/kubernetes/client-go/rest` provides
    `InClusterConfig`, a function that returns a client config utilising
    the service account assigned to the pod within which `alduin` would
    be running.

[^3]: See field `Image` of type `Container` in package `containers` in
    the `containerd` source code repository at
    `https://github.com/containerd/containerd`. A comment to the field
    `Name` of type `Image` in package `images` stating that an image
    name should be "a reference compatible with resolvers" seems to
    suggest delegation of decisions about name format.

[^4]: See example in section "Updating resources" of the `kubectl` Cheat
    Sheet at `https://kubernetes.io/docs/reference/kubectl/cheatsheet/`.
