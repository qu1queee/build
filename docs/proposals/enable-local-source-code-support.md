<!--
Copyright The Shipwright Contributors

SPDX-License-Identifier: Apache-2.0
-->

---
title: Parameterize Build Strategies
authors:
  - "@qu1queee"
reviewers:
  - "@SaschaSchwarze0"
  - "@ImJasonH"
  - "@sbose78"
  - "@zhangtbj"
  - "@xiujuan95"
approvers:
  - "@zhangtbj"
  - "@sbose78"
creation-date: 2021-04-05
last-updated: 2021-04-05
status: implementable
---

# Enable Local Source Code Support

## Release Signoff Checklist

- [x] Enhancement is `implementable`
- [x] Design details are appropriately documented from clear requirements
- [x] Test plan is defined
- [x] Graduation criteria for dev preview, tech preview, GA
- [ ] User-facing documentation is created in [docs](/docs/)

## Open Questions [optional]

> 1. TBD

TBD

> 2. TBD

TBD

## Summary

Enable Shipwright Users to build images from local source code, without the need of involving any Git repository service ( _e.g Github, Gitlab_ ). With the condition of having the appropriate runtime source code locally ( _in their local machine_ ). 

Enabling local source code requires a coordination of work between [Shipwright/Build](https://github.com/shipwright-io/build) and [Shipwright/CLI](https://github.com/shipwright-io/cli), where the first one exposes changes in the Build API, and the second one implements a bundle mechanism to pull/push the source code and embed it as a standalone step in the strategy of choice.

## Motivation

There are several reasons for the need of local source code:

- **Development Mode Support**: Provide developers the ability to build container images while working on development mode, without forcing them to _commit/push_ changes to their source code. This should also enable any workflow using Shipwright to provide a full end-to-end flow, where developers can build and deploy from their local development environment.

- **Closing The Gaps**: Local source code support inlines with the continuous effort on improving the developer experience. There are different vendors or projects that indirectly support this feature. A popular example is [Cloud Foundry](https://docs.cloudfoundry.org/devguide/deploy-apps/deploy-app.html) with the `cf push` experience, enabling developers to build and deploy any application from local source code. Also [Kf](https://cloud.google.com/migrate/kf/docs/2.3/quickstart) provides the same developer experience, but it runs in Kubernetes.

### Goals

- Fulfill a missing requirement in Tekton/Pipeline repository, which provides support for local source code when executing a series of Task steps.

- Enhance the Build resource API, by introducing new fields that signalize the usage of local source code. 

- Conclude on the best way to expose the usage of local source code via the Build API.

- Ensure local source code is transparent to users, in other words, we should not need to duplicate Shipwright strategies, this should be all done via a Build definition.

- Make the Build `spec.source.url` not mandatory.

- Layout a `bundle` mechanism on how to exploit the usage of a container registry, to push/pull local source code in combination with the Shipwright Strategies. The implementation of this mechanism should be done in Shipwright/CLI repo.

### Non-Goals

- Layout all different mechanisms to support local source code in Shipwright/CLI.

- Modify Shipwright strategies to support local source code.

- Define how to surface in the BuildRun status fields metadata about the local source code.

## Proposal

### Part 1: Build Resource API changes

This EP is well aware of all potential changes for the Build API in the future, such as:

- [Remote Artifacts](https://github.com/shipwright-io/build/blob/master/docs/proposals/remote-artifacts.md)
- [Build Inputs Overhaul](https://github.com/shipwright-io/build/pull/652)

But to be pragmatic and avoid confusion, this EP proposes to do the following changes in the current state of the Build API.

##### Modify the API

- Introduce `spec.source.local`, which should host the desire usage of this feature. The value of `spec.source.url` requires to be an absolute path to the code in the user local machine. See an [example](https://github.com/qu1queee/build/blob/qu1queee/local_source/pkg/apis/build/v1alpha1/source.go#L14-L16).

- Ensure `spec.source.url` is not longer a mandatory field. This was done thinking on assets only hosted in `git`, which no longer holds true. See an [example](https://github.com/qu1queee/build/blob/qu1queee/local_source/pkg/apis/build/v1alpha1/source.go#L19-L20).

#### Modify the Runtime logic

- Ensure that when `spec.source.local` is defined, we do not longer generate the Tekton Input PipelineResource. We do this today, to tell Tekton that we want to pull source from a git repository, which ends as a container that pulls it. See an [example](https://github.com/qu1queee/build/blob/qu1queee/local_source/pkg/reconciler/buildrun/resources/taskrun.go#L184-L185) on future changes.

- For the bundle mechanism, which is explained in the next section. We need to **prepend** a new step in our Task step definition, which will pull our local source code from a registry. See an [example](https://github.com/qu1queee/build/blob/qu1queee/local_source/pkg/reconciler/buildrun/resources/taskrun.go#L171-L186). 

  Important to notice, the image to pull will be a self-extracted image, therefore the `workingDir` container definition should be under `/workspace/source`, which is a well-known path in the Shipwright strategies, where source code is expected to be.

_Note_: More changes code-wise might be needed, but that should be an implementation detail and does not belong in this EP description.
_Note_: The above does not require a manual modification to any of the existing Shipwright strategies.


### Part 2: Bundle Mechanism

This is CLI specific. Any workflow using Shipwright that intends to provide support for local source code will require to re-use the Shipwright/CLI implementation or implement a standalone one.

This proposed implementation can be one of many. In the future we might want to use a different storage service as needed.

The Bundle mechanism has it´s origins from [Mink Bundles](https://github.com/mattmoor/mink/tree/master/pkg/bundles) and it´s based on the premise that if a user is pushing a container image to a container registry, then workflows could use the same approach to push the local source code, and pull it at a later stage during the building image mechanism.

The Bundle mechanism executes the following:

- [1] Takes the source code from a local directory.
- [2] Build a container image that can self-extracted on runtime.
- [3] Pushed the container image with the source code into a container registry, making it available for further usages.

In Shipwright/Build we will ensure that any strategy where local source code is required, can pull the bundle image [3] before calling any of the existing tooling, like `kaniko`, `paketo`, etc.

The current examples provided in this EP, re-use the existing packages from [Mink Bundles](https://github.com/mattmoor/mink/tree/master/pkg/bundles), but for Shipwright/CLI we require to have our own custom implementation for several reasons:

- Mink uses [go-containerregistry](https://github.com/google/go-containerregistry), so we should be able to do the same.
- Have control on the authentication part, we need to ensure authentication to push/pull the local source code can also reuse the secrets under `spec.output`.
- Have control on the base layers of the bundle image, which ensure that the image on `runtime`, can self-extract themself.

The following layouts some of the key points we will need:

- Introduce a new go pkg in the CLI side, to support the bundle approach. See an [example](https://github.com/qu1queee/cli/tree/qu1queee/crud_cmd/pkg/shp/bundle)
- Ensure that the `run` subcommand validates if local source code is desired. If this is the case, it should bundle and create an image. See an [example](https://github.com/qu1queee/cli/blob/qu1queee/crud_cmd/pkg/shp/cmd/build/run.go#L72-L90).
- Introduce any related subcommand to surface the usage of local source code.

### User Stories [optional]

Build users need to define the required parameter values if they want to opt-in for the usage of local source code feature.

#### As a Shipwright/Build contributor I want to have a well defined API to support local source code

The Build resource API needs to provide means of signalizing the desire of local source code. In terms of API changes this should be minimal.
This should also have some implications on previous assumptions, like:

- Remote Artifacts didnt consider the usage of local source code, how will this fit there?
- Git as a service for hosting repository cannot longer be mandatory, e.g. `spec.source.url` definition.

Therefore, users should be able to signalize this feature via:

```yaml
spec:
  source:
    local: /<path-to-local>/github.com/shipwright-io/sample-go/docker-build
```


#### As a Shipwright User I want to build images in Development mode from my local Dockerfile

```yaml
---
apiVersion: shipwright.io/v1alpha1
kind: Build
metadata:
  name: a-kaniko-build
spec:
  source:
    local: /<path-to-local>/github.com/shipwright-io/sample-go/docker-build
  strategy:
    name: kaniko
    kind: ClusterBuildStrategy
  output: ...
```

#### As a Shipwright User I want to build images in Development mode from my local Source code

```yaml
---
apiVersion: shipwright.io/v1alpha1
kind: Build
metadata:
  name: a-buildpacks-build
spec:
  source:
    local: /<path-to-local>/github.com/shipwright-io/sample-go/source-build
  strategy:
    name: buildpacks-v3
    kind: ClusterBuildStrategy
  output: ...
```


#### As a Shipwright/CLI contributor I want to provide a mechanism to support local source code

This is related to the CLI workflow, that should support the bundle approach and maintain the bundle base images. The implementation should be generic enough for other workflows, like OpenShift Build, IBM Cloud Code Engine, too reuse it.

### Implementation Details/Notes/Constraints [optional]

There is a prototype for educational purposes in:

- Changes in Shipwright/Build, see [branch](https://github.com/qu1queee/build/tree/qu1queee/local_source)
- Changes in Shipwright/CLI, see [branch](https://github.com/qu1queee/cli/tree/qu1queee/crud_cmd)

_Note_: The changes in CLI include a pinned [module](https://github.com/qu1queee/cli/blob/qu1queee/crud_cmd/go.mod#L23), please modify this to your local one.

### Risks and Mitigations

- This requires coordination between Build and CLI. A well defined list of issues for both backlogs should be created.

- Using a container registry to push/pull the source code might bring concerns around performance. This EP only proposes one approach on local source code. CLI should not lock-in with a single approach like bundles, but it should be flexible enough, to extend to new approaches in the future.

## Design Details

### Test Plan

- This requires new unit, integration and a new e2e tests with local source code.

### Graduation Criteria

Should be part of any release before our v1.0.0

### Upgrade / Downgrade Strategy

Not apply, this should not break anything, is rather an extension of the API.

### Version Skew Strategy

N/A

## Implementation History

N/A

## Drawbacks

None

## Alternatives

- None at the moment.