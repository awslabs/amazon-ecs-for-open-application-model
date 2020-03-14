// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package types

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/olekukonko/tablewriter"
)

// ComponentInput holds the fields required to deploy an component instance.
type ComponentInput struct {
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

func (component *Component) Display() {
	fmt.Printf("\nComponent Instance: %s\n\n", component.StackName)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Component Instance Attribute", "Value"})
	table.SetBorder(false)

	keys := make([]string, 0, len(component.StackOutputs))
	for key := range component.StackOutputs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		formattedKey := strings.Title(strings.ToLower(strcase.ToDelimited(key, ' ')))
		formattedKey = strings.ReplaceAll(formattedKey, "Cloud Formation", "CloudFormation")
		formattedKey = strings.ReplaceAll(formattedKey, "Ecs", "ECS")
		table.Append([]string{formattedKey, component.StackOutputs[key]})
	}

	table.Render()
	fmt.Println("")
}
