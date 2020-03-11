# Compatibility with Open Application Model

This document describes the support in oam-ecs for the Open Application Model specification.  This comparison is based on the OAM spec version `v1alpha1`, as of commit [4af9e65](https://github.com/oam-dev/spec/tree/4af9e65769759c408193445baf99eadd93f3426a).

Note that oam-ecs does very little validation to find OAM attributes it does not support in the given OAM files.  Most attributes that are not supported are silently ignored.

Legend:
* :heavy_check_mark: Full support
* :large_blue_diamond: Partial support
* :x: No support

## Top-Level Attributes

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/2.overview_and_terminology.md#representing-oam-objects-as-schematics)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `apiVersion` | Must be `core.oam.dev/v1alpha1` |
| :large_blue_diamond: | `kind` | `ApplicationConfiguration` and `ComponentSchematic` are supported. `Trait`, `WorkloadType`, and `ApplicationScope` are not supported. |
| :large_blue_diamond: | `metadata` | See [details below](#metadata) |
| :heavy_check_mark: | `spec` | |

### Metadata

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/2.overview_and_terminology.md#metadata)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `name` | |
| :x: | `labels` | Ignored |
| :x: | `annotations` | Ignored |

## Component Schematic Spec

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#spec)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `parameters` |  |
| :large_blue_diamond: | `workloadType` | See [details below](#workload-types) |
| :x: | `osType` | Ignored. `linux` is assumed. |
| :x: | `arch` | Ignored. `amd64` is assumed. |
| :large_blue_diamond: | `containers` | See [details below](#component-schematic-container) |
| :x: | `workloadSettings` | |

### Component Schematic Parameter

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#parameter)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `name` |  |
| :heavy_check_mark: | `description` |  |
| :large_blue_diamond: | `type` | Ignored. No validation occurs |
| :x: | `required` | |
| :x: | `default` | |

### Component Schematic Container

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#container)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `name` |  |
| :heavy_check_mark: | `image` |  |
| :large_blue_diamond: | `resources` | See [details below](#component-schematic-resources) |
| :heavy_check_mark: | `cmd` | |
| :heavy_check_mark: | `args` | |
| :heavy_check_mark: | `env` |  |
| :x: | `config` |  |
| :heavy_check_mark: | `ports` |  |
| :large_blue_diamond: | `livenessProbe` |  See [details below](#component-schematic-healthprobe) |
| :x: | `readinessProbe` | |
| :heavy_check_mark: | `imagePullSecret` | Must be the name (not ARN) of a Secrets Manager secret in the same region, encrypted with the default KMS key. See the [ECS documentation](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/private-auth.html) for instructions. |

### Component Schematic Resources

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#resources)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `cpu` | Will sum the CPU requirements of all containers in the workload, and round up to the best-fit [Fargate task size](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html) |
| :heavy_check_mark: | `memory` | Will sum the memory requirements of all containers in the workload, and round up to the best-fit [Fargate task size](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html) |
| :large_blue_diamond: | `gpu` | GPU requirements will be generated into the CloudFormation template, but the deployment will fail as Fargate does not support GPUs. |
| :large_blue_diamond: | `volumes` | Some volume requirements will be generated into the CloudFormation template, but the deployment will fail as Fargate does not support volumes. See [details below](#component-schematic-volume) for the supported volume requirements. |
| :x: | `extended` | |

### Component Schematic Volume

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#volume)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `name` |  |
| :heavy_check_mark: | `mountPath` |  |
| :heavy_check_mark: | `accessMode` |  |
| :x: | `sharingPolicy` | |
| :x: | `disk` | |

### Component Schematic HealthProbe

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#healthprobe)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `exec` |  |
| :large_blue_diamond: | `httpGet` | Only supported for workload type `core.oam.dev/v1alpha1.Server`. `httpHeaders` attribute is not supported. |
| :heavy_check_mark: | `tcpSocket` | Only supported for workload type `core.oam.dev/v1alpha1.Server` |
| :heavy_check_mark: | `initialDelaySeconds` | |
| :heavy_check_mark: | `periodSeconds` | |
| :heavy_check_mark: | `timeoutSeconds` | |
| :large_blue_diamond: | `successThreshold` | Not supported for `exec` health probe. |
| :heavy_check_mark: | `failureThreshold` | |

## Application Configuration Spec

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/6.application_configuration.md#spec)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :x: | `variables` |  |
| :x: | `scopes` |  |
| :heavy_check_mark: | `components` |  |

### Application Configuration Component

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/6.application_configuration.md#component)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `componentName` |  |
| :heavy_check_mark: | `instanceName` |  |
| :heavy_check_mark: | `parameterValues` |  |
| :heavy_check_mark: | `traits` |  |
| :x: | `applicationScopes` |  |

### Application Configuration Trait

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/6.application_configuration.md#trait)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `name` |  |
| :heavy_check_mark: | `properties` |  |

## Workload Types

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#workload-types)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `core.oam.dev/v1alpha1.Server` |  |
| :x: | `core.oam.dev/v1alpha1.SingletonServer` |  |
| :heavy_check_mark: | `core.oam.dev/v1alpha1.Worker` |  |
| :x: | `core.oam.dev/v1alpha1.SingletonWorker` |  |
| :x: | `core.oam.dev/v1alpha1.Task` | |
| :x: | `core.oam.dev/v1alpha1.SingletonTask` | |
| :x: | Extended workload types | |

## Application Scopes

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/4.application_scopes.md#application-scope-types)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :x: | `core.oam.dev/v1alpha1.Network` |  |
| :x: | `core.oam.dev/v1alpha1.Health` |  |
| :x: | Extended application scope types |  |

## Traits

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/5.traits.md#core-traits)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `manual-scaler` |  |
| :x: | Extended trait types |  |
