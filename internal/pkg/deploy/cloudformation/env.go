// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cloudformation provides functionality to deploy oam-ecs resources with AWS CloudFormation.
package cloudformation

import (
	"errors"

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

// DescribeEnvironment describes the existing CloudFormation stack for an environment
func (cf CloudFormation) DescribeEnvironment(env *types.EnvironmentInput) (*types.Environment, error) {
	envConfig := stack.NewEnvStackConfig(env, cf.box)
	stack, err := cf.describe(envConfig)
	if err != nil {
		return nil, err
	}
	return envConfig.ToEnv(stack)
}
