// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cli contains the oam-ecs subcommands.
package cli

import (
	"fmt"

	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/aws/session"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation/types"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/log"
	termprogress "github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/progress"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/workload"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/spf13/cobra"
)

const (
	showComponentStart     = "Retrieving the infrastructure information for the component instance %s."
	showComponentFailed    = "Failed to retrieve the infrastructure information for the component instance %s."
	showComponentSucceeded = "Retrieved the infrastructure information for component instance %s in CloudFormation stack %s."
)

type cfComponentDescriber interface {
	DescribeComponent(component *types.ComponentInput) (*types.Component, error)
}

// ShowAppOpts holds the configuration needed to describe an application.
type ShowAppOpts struct {
	// Fields with matching flags
	OamFile string

	prog               progress
	ComponentDescriber cfComponentDescriber
}

// NewShowAppOpts initiates the fields to describe an application.
func NewShowAppOpts() *ShowAppOpts {
	return &ShowAppOpts{
		prog: termprogress.NewSpinner(),
	}
}

func (opts *ShowAppOpts) newComponentInput(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration) (*types.ComponentInput, error) {
	environment := &types.ComponentEnvironment{
		Name: environmentName,
	}

	return &types.ComponentInput{
		ApplicationConfiguration: application,
		ComponentConfiguration:   componentInstance,
		Environment:              environment,
	}, nil
}

func (opts *ShowAppOpts) showComponentInstance(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration) error {
	componentInput, err := opts.newComponentInput(application, componentInstance)
	if err != nil {
		return err
	}

	opts.prog.Start(fmt.Sprintf(showComponentStart, componentInstance.InstanceName))

	component, err := opts.ComponentDescriber.DescribeComponent(componentInput)
	if err != nil {
		opts.prog.Stop(log.Serrorf(showComponentFailed, componentInstance.InstanceName))
		return err
	}

	opts.prog.Stop(log.Ssuccessf(showComponentSucceeded, componentInstance.InstanceName, component.StackName))

	component.Display()

	return nil
}

// Execute parses the OAM files and shows the infrastructure for the application configuration
func (opts *ShowAppOpts) Execute() error {
	oamWorkload, err := workload.NewOamWorkload(
		&workload.OamWorkloadProps{
			OamFiles: []string{opts.OamFile},
		})
	if err != nil {
		return err
	}

	// Show the application components
	for _, componentInstance := range oamWorkload.ApplicationConfiguration.Spec.Components {
		err = opts.showComponentInstance(oamWorkload.ApplicationConfiguration, &componentInstance)

		if err != nil {
			break
		}
	}

	return err
}

// BuildShowAppCmd build the command for showing an application.
func BuildShowAppCmd() *cobra.Command {
	opts := NewShowAppOpts()
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Describe the application's deployed components",
		Long:  `Retrieves and displays the attributes of the infrastructure for the application defined in an Open Application Model application configuration file.`,
		Example: `
  Show the deployed application components, using an application configuration file:
	$ oam-ecs app show -f config.yml`,
		PreRunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			session, err := session.Default()
			if err != nil {
				return err
			}
			opts.ComponentDescriber = cloudformation.New(session)
			return nil
		}),
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			return opts.Execute()
		}),
	}

	cmd.Flags().StringVarP(&opts.OamFile, oamFileFlag, oamFileFlagShort, "", appConfigFileFlagDescription)
	cmd.MarkFlagRequired(oamFileFlag)

	return cmd
}
