// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package types

import "github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"

// DeployComponentInput holds the fields required to deploy an component instance.
type DeployComponentInput struct {
	ApplicationConfiguration *v1alpha1.ApplicationConfiguration
	ComponentConfiguration   *v1alpha1.ComponentConfiguration
	Component                *v1alpha1.ComponentSchematic
	Environment              *ComponentEnvironment
	WorkloadSettings         *ECSWorkloadSettings
}

// ECSWorkloadSettings holds fields that are needed to define services in ECS, which are not part of the core OAM types
type ECSWorkloadSettings struct {
	TaskCPU    string
	TaskMemory string
}

// Environment represents attributes about the environment where the component will be deployed
type ComponentEnvironment struct {
	Name string
}

// Component represents the configuration of a particular component instance
type Component struct {
	StackName    string
	StackOutputs map[string]string
}

// DeployComponenttResponse holds the created component instance on successful deployment.
// Otherwise, the component is set to nil and a descriptive error is returned.
type DeployComponentResponse struct {
	Component *Component
	Err       error
}
