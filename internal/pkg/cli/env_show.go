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
	showEnvStart     = "Retrieving the infrastructure information for the environment."
	showEnvFailed    = "Failed to retrieve the infrastructure information for the environment."
	showEnvSucceeded = "Retrieved the infrastructure information for CloudFormation stack %s."
)

type cfEnvironmentDescriber interface {
	DescribeEnvironment(env *types.EnvironmentInput) (*types.Environment, error)
}

type ShowEnvironmentOpts struct {
	prog         progress
	envDescriber cfEnvironmentDescriber
}

func NewShowEnvironmentOpts() *ShowEnvironmentOpts {
	return &ShowEnvironmentOpts{
		prog: termprogress.NewSpinner(),
	}
}

func (opts *ShowEnvironmentOpts) Execute() error {
	describeEnvInput := &types.EnvironmentInput{}

	opts.prog.Start(showEnvStart)

	env, err := opts.envDescriber.DescribeEnvironment(describeEnvInput)
	if err != nil {
		opts.prog.Stop(log.Serror(showEnvFailed))
		return err
	}

	opts.prog.Stop(log.Ssuccessf(showEnvSucceeded, env.StackName))

	env.Display()

	return nil
}

func BuildShowEnvironmentCmd() *cobra.Command {
	opts := NewShowEnvironmentOpts()
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Describe the oam-ecs environment",
		Long:  `Retrieves and displays the attributes of the oam-ecs default environment`,
		Example: `
  Show the oam-ecs environment:
	$ oam-ecs env show`,
		PreRunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			session, err := session.Default()
			if err != nil {
				return err
			}
			opts.envDescriber = cloudformation.New(session)
			return nil
		}),
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			return opts.Execute()
		}),
	}

	return cmd
}
