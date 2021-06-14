<!--
Copyright The Shipwright Contributors

SPDX-License-Identifier: Apache-2.0
-->

# Building a Strategy

Before starting, make sure you have Tekton and Shipwright Build installed.

See the [Try It!](../../README.md#try-it) section for more information.

## Getting Started

### Registry Authentication

For this tutorial, we will require to create a `tutorial-secret` Kubernetes secret to access a [DockerHub](https://hub.docker.com/) registry, as follows:

```sh
$ REGISTRY_SERVER=https://index.docker.io/v1/ REGISTRY_USER=<your_registry_user> REGISTRY_PASSWORD=<your_registry_password>
$ kubectl create secret docker-registry tutorial-secret --docker-server=$REGISTRY_SERVER --docker-username=$REGISTRY_USER --docker-password=$REGISTRY_PASSWORD  --docker-email=me@here.com
```

_Note_: For more information about authentication, please refer to the related [docs](/docs/development/authentication.md).

## Strategy Concepts

In the following sections, the fundamental concepts around strategies are explained. These concepts will allow users to construct optimal strategies.

### **Personas**

In Shipwright, we manage two types of personas. See table:

| Persona                       | Role                                                           |
| ----------------------------- | -------------------------------------------------------------- |
| **Strategy Authors**              | They define, install and maintain Build Strategies resources in the clusters                             |
| **Build Users**                   | Define Build and BuildRun resources. They should reference a strategy to use in their Build Definitions        |

Strategy Authors are responsible for providing answers to the following questions when building strategies:

- _What is the cluster-scope I want to have in the strategy?_
- _How many steps I want to have in the strategy? and why?_
- _What are the available system parameters and how can I use them?_
- _Should I defined custom strategy parameters for Build Users?_

### **Scope**

A Strategy scope can be of the type:

- **cluster-scope:** Available to Builds in all namespaces.
- **namespaced-scope:** Available only to Builds in a particular namespace.

### **Steps**

A strategy step is analogous to Kubernetes containers [API](https://github.com/kubernetes/api/blob/master/core/v1/types.go#L2205). A step contains everything needed to run an application code. On a step definition, one must define at least:

- The container `image` to use.
- The container `command` and `arguments` to execute.

But also, strategy authors can define:

- The container `securityContext`.
- The container `imagePullPolicy`.
- The container `resources`.

Note that, a single Strategy will result on a single Kubernetes pod running N number of containers ( _steps_ ). In Shipwright, we try to keep the number of steps per strategy to a minimum ( _reducing resources consumption_ ).

### **Tooling**

The tooling/binaries ( _applications_ ) we use on a _step_ basis is focus on the following:

- **Pulling source code:** Tooling that allow us to pull source code from _git_ . See our custom [application](https://quay.io/repository/shipwright/git)

- **Images builders:** Tooling that allow us to build container images out of source code. We currently have multiple options depending on the need:
  - [Buildpacks-v3](docs/buildstrategies.md#buildpacks-v3)
  - [Kaniko](docs/buildstrategies.md#kaniko)
  - [BuildKit](docs/buildstrategies.md#buildkit)
  - [Source-to-Image](docs/buildstrategies.md#source-to-image)
  - [Buildah](docs/buildstrategies.md#buildah)
  - [ko](docs/buildstrategies.md#ko)

For **Images Builders** is important to understand what is the requirement. For example, do you want to build a container image using a _Dockerfile_, or without a _Dockerfile_.

### **System Parameters**

These are parameters that are by default available to all strategy authors to use on their Strategies definition. The following highlights the available system parameters:

| Parameter                      | Description |
| ------------------------------ | ----------- |
| shp-source-root    | The absolute path to the directory that contains the Build User's sources. |
| shp-source-context | The absolute path to the context directory of the Build User's sources. If the user specified no value for `spec.source.contextDir` in their Build, then this value will equal the value for `shp-source-root` |
| shp-output-image      | The URL of the image that the user wants to push as specified in the Build's `spec.output.image`, or the override from the BuildRun's `spec.output.image`. |

For using system parameters on the strategies, we follow the Tekton [notation](https://github.com/tektoncd/pipeline/blob/main/docs/tasks.md#using-variable-substitution) for parameters. As follows:

To reference a parameter in a Strategy step, use the following syntax, where `<name>` is the name of the system parameter:

```sh
$(params.<name>)
```

### Strategy Parameters

These are parameters that are defined by strategy authors, and are intended to be modify by Build Users on the related Build or BuildRun definition.

A Strategy parameter consist of:

- **name**: The Name of the parameter.
- **description**: A proper description that highlights the usage intention of the parameter.
- **default**: A reasonable default value ( _optional_ ).

To define a parameter in the strategy, this needs to be defined under `spec.parameters`. Build Users will be able to control the values of these parameters by defining the same parameter name under their `spec.params` object.

To reference a parameter in a Strategy step, use the following syntax, where `<name>` is the name of the strategy parameter:

```sh
$(params.<name>)
```

## Creating a Strategy
