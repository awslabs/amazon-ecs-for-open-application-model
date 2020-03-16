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

// DeployEnvironment creates the CloudFormation stack for an environment by creating and executing a change set.
//
// If the deployment succeeds, returns nil.
// If the stack already exists, update the stack.
// If the change set to create/update the stack cannot be executed, returns a ErrNotExecutableChangeSet.
// Otherwise, returns a wrapped error.
func (cf CloudFormation) DeployEnvironment(env *types.EnvironmentInput) (*types.Environment, error) {
	envConfig := stack.NewEnvStackConfig(env, cf.box)

	// Try to create the stack
	if _, err := cf.create(envConfig); err != nil {
		var existsErr *ErrStackAlreadyExists
		if errors.As(err, &existsErr) {
			// Stack already exists, update the stack
			deployStarted, err := cf.update(envConfig)
			if err != nil {
				return nil, err
			}

			if deployStarted {
				// Wait for the stack to finish updating
				stack, err := cf.waitForStackUpdate(envConfig)
				if err != nil {
					return nil, err
				}
				return envConfig.ToEnv(stack)
			} else {
				// nothing to deploy
				stack, err := cf.describe(envConfig)
				if err != nil {
					return nil, err
				}
				return envConfig.ToEnv(stack)
			}
		} else {
			return nil, err
		}
	}

	// Wait for the stack to finish creation
	stack, err := cf.waitForStackCreation(envConfig)
	if err != nil {
		return nil, err
	}
	return envConfig.ToEnv(stack)
}

func (cf CloudFormation) DryRunEnvironment(env *types.EnvironmentInput) (string, error) {
	stackConfig := stack.NewEnvStackConfig(env, cf.box)
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

// DescribeEnvironment describes the existing CloudFormation stack for an environment
func (cf CloudFormation) DescribeEnvironment(env *types.EnvironmentInput) (*types.Environment, error) {
	envConfig := stack.NewEnvStackConfig(env, cf.box)
	stack, err := cf.describe(envConfig)
	if err != nil {
		return nil, err
	}
	return envConfig.ToEnv(stack)
}

// DeleteEnvironment deletes the CloudFormation stack for an environment
func (cf CloudFormation) DeleteEnvironment(env *types.EnvironmentInput) (*types.Environment, error) {
	envConfig := stack.NewEnvStackConfig(env, cf.box)
	stack, err := cf.describe(envConfig)
	if err != nil {
		var notFoundErr *ErrStackNotFound
		if errors.As(err, &notFoundErr) {
			// Stack was not found, don't return an error, since it's deleted already
			return &types.Environment{
				StackName: envConfig.StackName(),
			}, nil
		} else {
			return nil, err
		}
	}
	err = cf.delete(*stack.StackId)
	if err != nil {
		return nil, err
	}
	return envConfig.ToEnv(stack)
}
