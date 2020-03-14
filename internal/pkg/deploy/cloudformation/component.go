// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cloudformation provides functionality to deploy oam-ecs resources with AWS CloudFormation.
package cloudformation

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation/stack"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation/types"
)

const (
	templateFileDirectoryName = "oam-ecs-dry-run-results"
)

// DeployComponent creates the CloudFormation stack for a component instance by creating and executing a change set.
//
// If the deployment succeeds, returns nil.
// If the stack already exists, update the stack.
// If the change set to create/update the stack cannot be executed, returns a ErrNotExecutableChangeSet.
// Otherwise, returns a wrapped error.
func (cf CloudFormation) DeployComponent(component *types.ComponentInput) (*types.Component, error) {
	componentConfig := stack.NewComponentStackConfig(component, cf.box)

	// Try to create the stack
	if _, err := cf.create(componentConfig); err != nil {
		var existsErr *ErrStackAlreadyExists
		if errors.As(err, &existsErr) {
			// Stack already exists, update the stack
			deployStarted, err := cf.update(componentConfig)
			if err != nil {
				return nil, err
			}

			if deployStarted {
				// Wait for the stack to finish updating
				stack, err := cf.waitForStackUpdate(componentConfig)
				if err != nil {
					return nil, err
				}
				return componentConfig.ToComponent(stack)
			} else {
				// nothing to deploy
				stack, err := cf.describe(componentConfig)
				if err != nil {
					return nil, err
				}
				return componentConfig.ToComponent(stack)
			}
		} else {
			return nil, err
		}
	}

	// Wait for the stack to finish creation
	stack, err := cf.waitForStackCreation(componentConfig)
	if err != nil {
		return nil, err
	}
	return componentConfig.ToComponent(stack)
}

func (cf CloudFormation) DryRunComponent(component *types.ComponentInput) (string, error) {
	stackConfig := stack.NewComponentStackConfig(component, cf.box)
	template, err := stackConfig.Template()
	if err != nil {
		return "", fmt.Errorf("template creation: %w", err)
	}

	templateFileDir := filepath.Join(".", templateFileDirectoryName)
	if _, err := os.Stat(templateFileDir); os.IsNotExist(err) {
		err = os.Mkdir(templateFileDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("could not create directory %s: %w", templateFileDir, err)
		}
	}

	templateFileAbsDir, err := filepath.Abs(templateFileDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for directory %s: %w", templateFileDir, err)
	}

	templateFilePath := filepath.Join(templateFileAbsDir, stackConfig.StackName()+"-template.yaml")

	f, err := os.Create(templateFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = f.WriteString(template)
	if err != nil {
		return "", err
	}
	f.Sync()

	return templateFilePath, nil
}

// DescribeComponent describes the existing CloudFormation stack for a component instance
func (cf CloudFormation) DescribeComponent(component *types.ComponentInput) (*types.Component, error) {
	stackConfig := stack.NewComponentStackConfig(component, cf.box)
	stack, err := cf.describe(stackConfig)
	if err != nil {
		return nil, err
	}
	return stackConfig.ToComponent(stack)
}

// DeleteComponent deletes the CloudFormation stack for a component instance
func (cf CloudFormation) DeleteComponent(component *types.ComponentInput) (*types.Component, error) {
	stackConfig := stack.NewComponentStackConfig(component, cf.box)
	stack, err := cf.describe(stackConfig)
	if err != nil {
		var notFoundErr *ErrStackNotFound
		if errors.As(err, &notFoundErr) {
			// Stack was not found, don't return an error, since it's deleted already
			return &types.Component{
				StackName: stackConfig.StackName(),
			}, nil
		} else {
			return nil, err
		}
	}
	err = cf.delete(*stack.StackId)
	if err != nil {
		return nil, err
	}
	return stackConfig.ToComponent(stack)
}
