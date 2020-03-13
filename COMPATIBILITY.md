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
| :heavy_check_mark: | `name` | CloudFormation stack name will be `oam-ecs-{application configuration name}-{component instance name}` |
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
| :x: | `required` | Ignored. No validation occurs |
| :heavy_check_mark: | `default` | |

### Component Schematic Container

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#container)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `name` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition Name](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-name) |
| :heavy_check_mark: | `image` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition Image](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-image) |
| :large_blue_diamond: | `resources` | See [details below](#component-schematic-resources) |
| :heavy_check_mark: | `cmd` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition EntryPoint](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-entrypoint) |
| :heavy_check_mark: | `args` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition Command](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-command) |
| :heavy_check_mark: | `env` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition Environment](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-environment) |
| :x: | `config` | ECS does not natively support config files |
| :heavy_check_mark: | `ports` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition PortMappings](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-portmappings), [AWS::ECS::Service LoadBalancers ContainerPort](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-service-loadbalancers.html#cfn-ecs-service-loadbalancers-containerport), [AWS::ElasticLoadBalancingV2::TargetGroup](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html) Port and Protocol, and [AWS::ElasticLoadBalancingV2::Listener](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-listener.html) Port and Protocol |
| :large_blue_diamond: | `livenessProbe` |  See [details below](#component-schematic-healthprobe) |
| :x: | `readinessProbe` | ECS does not distinguish between liveness and readiness |
| :heavy_check_mark: | `imagePullSecret` | Translates to [AWS::ECS::TaskDefinition RepositoryCredentials](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-repositorycredentials).<br>Must be the name (not ARN) of a Secrets Manager secret in the same region, encrypted with the default KMS key. See the [ECS documentation](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/private-auth.html) for instructions |

### Component Schematic Resources

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#resources)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `cpu` | Translates to [AWS::ECS::TaskDefinition Cpu](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-taskdefinition.html#cfn-ecs-taskdefinition-cpu).<br>Will sum the CPU requirements of all containers in the workload, and round up to the best-fit [Fargate task size](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html) |
| :heavy_check_mark: | `memory` | Translates to [AWS::ECS::TaskDefinition Memory](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-taskdefinition.html#cfn-ecs-taskdefinition-memory).<br>Will sum the memory requirements of all containers in the workload, and round up to the best-fit [Fargate task size](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html) |
| :large_blue_diamond: | `gpu` | Translates to [AWS::ECS::TaskDefinition ResourceRequirements](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-taskdefinition.html#cfn-ecs-taskdefinition-resourcerequirements).<br>GPU requirements will be generated into the CloudFormation template, but the deployment will fail as Fargate does not support GPUs. |
| :large_blue_diamond: | `volumes` | Some volume requirements will be generated into the CloudFormation template, but the deployment will fail as Fargate does not support volumes. See [details below](#component-schematic-volume) for the supported volume requirements. |
| :x: | `extended` | |

### Component Schematic Volume

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#volume)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `name` | Translates to [AWS::ECS::TaskDefinition Volumes Name](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-volumes.html#cfn-ecs-taskdefinition-volumes-name) and [AWS::ECS::TaskDefinition ContainerDefinition MountPoints Name](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-mountpoints.html#cfn-ecs-taskdefinition-containerdefinition-mountpoints-sourcevolume) |
| :heavy_check_mark: | `mountPath` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition MountPoints ContainerPath](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-mountpoints.html#cfn-ecs-taskdefinition-containerdefinition-mountpoints-containerpath) |
| :heavy_check_mark: | `accessMode` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition MountPoints ReadOnly](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions-mountpoints.html#cfn-ecs-taskdefinition-containerdefinition-mountpoints-readonly) |
| :x: | `sharingPolicy` | |
| :x: | `disk` | |

### Component Schematic HealthProbe

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#healthprobe)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `exec` | Translates to [AWS::ECS::TaskDefinition ContainerDefinition HealthCheck Command](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-healthcheck.html#cfn-ecs-taskdefinition-healthcheck-command) |
| :large_blue_diamond: | `httpGet` | Only supported for workload type `core.oam.dev/v1alpha1.Server`. `httpHeaders` attribute is not supported. Translates to [AWS::ElasticLoadBalancingV2::TargetGroup](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html) HealthCheckPath, HealthCheckPort, and HealthCheckProtocol |
| :heavy_check_mark: | `tcpSocket` | Only supported for workload type `core.oam.dev/v1alpha1.Server`. Translates to [AWS::ElasticLoadBalancingV2::TargetGroup](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html) HealthCheckPort and HealthCheckProtocol |
| :heavy_check_mark: | `initialDelaySeconds` | With `exec`, translates to [AWS::ECS::TaskDefinition ContainerDefinition HealthCheck StartPeriod](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-healthcheck.html#cfn-ecs-taskdefinition-healthcheck-startperiod).<br>With `httpGet` or `tcpSocket`, translates to [AWS::ECS::Service HealthCheckGracePeriodSeconds](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-service.html#cfn-ecs-service-healthcheckgraceperiodseconds) |
| :heavy_check_mark: | `periodSeconds` | With `exec`, translates to [AWS::ECS::TaskDefinition ContainerDefinition HealthCheck Interval](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-healthcheck.html#cfn-ecs-taskdefinition-healthcheck-interval).<br>With `httpGet` or `tcpSocket`, translates to [AWS::ElasticLoadBalancingV2::TargetGroup HealthCheckIntervalSeconds](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html#cfn-elasticloadbalancingv2-targetgroup-healthcheckintervalseconds) |
| :heavy_check_mark: | `timeoutSeconds` | With `exec`, translates to [AWS::ECS::TaskDefinition ContainerDefinition HealthCheck Timeout](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-healthcheck.html#cfn-ecs-taskdefinition-healthcheck-timeout). Defaults to 2, instead of the OAM spec default of 1.<br>With `httpGet` or `tcpSocket`, translates to [AWS::ElasticLoadBalancingV2::TargetGroup HealthCheckTimeoutSeconds](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html#cfn-elasticloadbalancingv2-targetgroup-healthchecktimeoutseconds). For `httpGet`, defaults to 6. For `tcpSocket`, defaults to 10. |
| :large_blue_diamond: | `successThreshold` | Not supported for `exec` health probe.<br>With `httpGet` or `tcpSocket`, translates to [AWS::ElasticLoadBalancingV2::TargetGroup HealthyThresholdCount](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html#cfn-elasticloadbalancingv2-targetgroup-healthythresholdcount). Defaults to 2, instead of the OAM spec default of 1. |
| :heavy_check_mark: | `failureThreshold` | With `exec`, translates to [AWS::ECS::TaskDefinition ContainerDefinition HealthCheck Retries](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-healthcheck.html#cfn-ecs-taskdefinition-healthcheck-retries).<br>With `httpGet` or `tcpSocket`, translates to [AWS::ElasticLoadBalancingV2::TargetGroup UnhealthyThresholdCount](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-elasticloadbalancingv2-targetgroup.html#cfn-elasticloadbalancingv2-targetgroup-unhealthythresholdcount) |

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
| :heavy_check_mark: | `instanceName` | CloudFormation stack name will be `oam-ecs-{application configuration name}-{component instance name}` |
| :heavy_check_mark: | `parameterValues` |  |
| :large_blue_diamond: | `traits` | See [details below](#traits) |
| :x: | `applicationScopes` |  |

## Workload Types

[(spec link)](https://github.com/oam-dev/spec/blob/4af9e65769759c408193445baf99eadd93f3426a/3.component_model.md#workload-types)

| Support | Attribute | Notes |
|---------|-----------|-------|
| :heavy_check_mark: | `core.oam.dev/v1alpha1.Server` | Translates to an ECS service running on Fargate, behind a Network Load Balancer |
| :x: | `core.oam.dev/v1alpha1.SingletonServer` |  |
| :heavy_check_mark: | `core.oam.dev/v1alpha1.Worker` | Translates to an ECS service running on Fargate, with no accessible endpoint |
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
| :heavy_check_mark: | `manual-scaler` | Translates to [AWS::ECS::Service DesiredCount](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ecs-service.html#cfn-ecs-service-desiredcount) |
| :x: | Extended trait types |  |
