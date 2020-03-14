// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package cli

import (
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/aws/session"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation/types"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/log"
	termprogress "github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/progress"
	"github.com/spf13/cobra"
)

const (
	deleteEnvStart     = "Deleting the infrastructure for the environment."
	deleteEnvFailed    = "Failed to delete the infrastructure for the environment."
	deleteEnvSucceeded = "Deleted the environment infrastructure in CloudFormation stack %s."
)

type cfEnvironmentDeleter interface {
	DeleteEnvironment(env *types.EnvironmentInput) (*types.Environment, error)
}

// DeleteEnvironmentOpts holds the configuration needed to deletes the oam-ecs environment.
type DeleteEnvironmentOpts struct {
	prog       progress
	envDeleter cfEnvironmentDeleter
}

// DeleteEnvironmentOpts initiates the fields to provision an environment.
func NewDeleteEnvironmentOpts() *DeleteEnvironmentOpts {
	return &DeleteEnvironmentOpts{
		prog: termprogress.NewSpinner(),
	}
}

// Execute deletes the environment CloudFormation stack
func (opts *DeleteEnvironmentOpts) Execute() error {
	deleteEnvInput := &types.EnvironmentInput{}

	opts.prog.Start(deleteEnvStart)

	env, err := opts.envDeleter.DeleteEnvironment(deleteEnvInput)
	if err != nil {
		opts.prog.Stop(log.Serror(deleteEnvFailed))
		return err
	}

	opts.prog.Stop(log.Ssuccessf(deleteEnvSucceeded, env.StackName))

	return nil
}

// BuildDeleteEnvironmentCmd build the command for creating a new pipeline.
func BuildDeleteEnvironmentCmd() *cobra.Command {
	opts := NewDeleteEnvironmentOpts()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the oam-ecs environment",
		Long:  `Removes the shared infrastructure, including a VPC and ECS cluster, for oam-ecs applications.  All deployed components must already be deleted.`,
		Example: `
  Delete the oam-ecs environment:
	$ oam-ecs env delete`,
		PreRunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			session, err := session.Default()
			if err != nil {
				return err
			}
			opts.envDeleter = cloudformation.New(session)
			return nil
		}),
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			return opts.Execute()
		}),
	}

	return cmd
}
