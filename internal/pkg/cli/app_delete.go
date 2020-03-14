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
	deleteComponentStart     = "Deleting the infrastructure for the component instance %s."
	deleteComponentFailed    = "Failed to delete the infrastructure for the component instance %s."
	deleteComponentSucceeded = "Deleted the infrastructure for component instance %s in CloudFormation stack %s."
)

type cfComponentDeleter interface {
	DeleteComponent(component *types.ComponentInput) (*types.Component, error)
}

// DeleteAppOpts holds the configuration needed to delete an application.
type DeleteAppOpts struct {
	// Fields with matching flags
	OamFile string

	prog             progress
	ComponentDeleter cfComponentDeleter
}

// NewDeleteAppOpts initiates the fields to delete an application.
func NewDeleteAppOpts() *DeleteAppOpts {
	return &DeleteAppOpts{
		prog: termprogress.NewSpinner(),
	}
}

func (opts *DeleteAppOpts) newComponentInput(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration) (*types.ComponentInput, error) {
	environment := &types.ComponentEnvironment{
		Name: environmentName,
	}

	return &types.ComponentInput{
		ApplicationConfiguration: application,
		ComponentConfiguration:   componentInstance,
		Environment:              environment,
	}, nil
}

func (opts *DeleteAppOpts) deleteComponentInstance(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration) error {
	componentInput, err := opts.newComponentInput(application, componentInstance)
	if err != nil {
		return err
	}

	opts.prog.Start(fmt.Sprintf(deleteComponentStart, componentInstance.InstanceName))

	component, err := opts.ComponentDeleter.DeleteComponent(componentInput)
	if err != nil {
		opts.prog.Stop(log.Serrorf(deleteComponentFailed, componentInstance.InstanceName))
		return err
	}

	opts.prog.Stop(log.Ssuccessf(deleteComponentSucceeded, componentInstance.InstanceName, component.StackName))

	return nil
}

// Execute parses the OAM files and deletes the infrastructure for the application configuration
func (opts *DeleteAppOpts) Execute() error {
	oamWorkload, err := workload.NewOamWorkload(
		&workload.OamWorkloadProps{
			OamFiles: []string{opts.OamFile},
		})
	if err != nil {
		return err
	}

	// Delete the application components
	for _, componentInstance := range oamWorkload.ApplicationConfiguration.Spec.Components {
		err = opts.deleteComponentInstance(oamWorkload.ApplicationConfiguration, &componentInstance)

		if err != nil {
			break
		}
	}

	return err
}

// BuildDeleteAppCmd builds the command for deleting an application.
func BuildDeleteAppCmd() *cobra.Command {
	opts := NewDeleteAppOpts()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete the application's deployed components",
		Long:  `Removes the infrastructure for the application defined in an Open Application Model application configuration file.`,
		Example: `
  Delete the deployed application components, using an application configuration file:
	$ oam-ecs app delete -f config.yml`,
		PreRunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			session, err := session.Default()
			if err != nil {
				return err
			}
			opts.ComponentDeleter = cloudformation.New(session)
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
