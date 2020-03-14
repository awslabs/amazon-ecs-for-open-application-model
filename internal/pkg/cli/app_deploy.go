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
	environmentName = "oam-ecs"

	dryRunComponentStart     = "Computing infrastructure template for the component instance %s."
	dryRunComponentFailed    = "Failed to generate infrastructure template for the component instance %s."
	dryRunComponentSucceeded = "Wrote infrastructure template to disk for component instance %s: %s"
	deployComponentStart     = "Deploying infrastructure changes for the component instance %s."
	deployComponentFailed    = "Failed to deploy infrastructure changes for the component instance %s."
	deployComponentSucceeded = "Deployed component instance %s in CloudFormation stack %s."
)

type cfComponentDeployer interface {
	DeployComponent(component *types.ComponentInput) (*types.Component, error)
	DryRunComponent(component *types.ComponentInput) (string, error)
}

// DeployAppOpts holds the configuration needed to provision an application.
type DeployAppOpts struct {
	// Fields with matching flags
	OamFiles []string
	DryRun   bool

	prog              progress
	ComponentDeployer cfComponentDeployer
}

// NewDeployAppOpts initiates the fields to provision an application.
func NewDeployAppOpts() *DeployAppOpts {
	return &DeployAppOpts{
		prog: termprogress.NewSpinner(),
	}
}

func (opts *DeployAppOpts) newComponentInput(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration, schematic *v1alpha1.ComponentSchematic) (*types.ComponentInput, error) {
	// TODO validate that following are not set: osType, arch, volume disk, volume sharing policy,
	// 				container extended resource, container config file, container readiness probe,
	//				container liveness probe failure threshold/httpGet/tcpSocket

	ecsSettings := &types.ECSWorkloadSettings{}

	environment := &types.ComponentEnvironment{
		Name: environmentName,
	}

	return &types.ComponentInput{
		ApplicationConfiguration: application,
		ComponentConfiguration:   componentInstance,
		Component:                schematic,
		WorkloadSettings:         ecsSettings,
		Environment:              environment,
	}, nil
}

func (opts *DeployAppOpts) dryRunComponentInstance(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration, schematic *v1alpha1.ComponentSchematic) error {
	deployComponentInput, err := opts.newComponentInput(application, componentInstance, schematic)
	if err != nil {
		return err
	}

	file, err := opts.ComponentDeployer.DryRunComponent(deployComponentInput)
	if err != nil {
		return err
	}

	log.Successln(fmt.Sprintf(dryRunComponentSucceeded, componentInstance.InstanceName, file))

	return nil
}

func (opts *DeployAppOpts) deployComponentInstance(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration, schematic *v1alpha1.ComponentSchematic) error {
	deployComponentInput, err := opts.newComponentInput(application, componentInstance, schematic)
	if err != nil {
		return err
	}

	opts.prog.Start(fmt.Sprintf(deployComponentStart, componentInstance.InstanceName))

	component, err := opts.ComponentDeployer.DeployComponent(deployComponentInput)
	if err != nil {
		opts.prog.Stop(log.Serrorf(deployComponentFailed, componentInstance.InstanceName))
		return err
	}

	opts.prog.Stop(log.Ssuccessf(deployComponentSucceeded, componentInstance.InstanceName, component.StackName))

	component.Display()

	return nil
}

// Execute parses the OAM files, translates them into infrastructure definitions, and deploys the infrastructure
func (opts *DeployAppOpts) Execute() error {
	oamWorkload, err := workload.NewOamWorkload(
		&workload.OamWorkloadProps{
			OamFiles: opts.OamFiles,
		})
	if err != nil {
		return err
	}

	// Validate we have app config and component schematics that go together
	for _, component := range oamWorkload.ApplicationConfiguration.Spec.Components {
		_, ok := oamWorkload.ComponentSchematics[component.ComponentName]
		if !ok {
			log.Errorf("Could not find the component schematic for %s\n", component.ComponentName)
			return fmt.Errorf("Application configuration refers to component %s, but no file provided the component schematic", component.ComponentName)
		}
	}

	// Deploy or dry-run the application components
	for _, componentInstance := range oamWorkload.ApplicationConfiguration.Spec.Components {
		schematic, _ := oamWorkload.ComponentSchematics[componentInstance.ComponentName]
		if opts.DryRun {
			err = opts.dryRunComponentInstance(oamWorkload.ApplicationConfiguration, &componentInstance, schematic)
		} else {
			err = opts.deployComponentInstance(oamWorkload.ApplicationConfiguration, &componentInstance, schematic)
		}

		if err != nil {
			break
		}
	}

	return err
}

// BuildDeployAppCmd build the command for deploying an application.
func BuildDeployAppCmd() *cobra.Command {
	opts := NewDeployAppOpts()
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the application",
		Long:  `Provisions (or updates) the Amazon ECS infrastructure for the application defined using the Open Application Model spec. All component schematics and the application configuration file for the application must be provided every time the 'app deploy' command runs (this CLI does not save any state).`,
		Example: `
  Deploy the application's OAM component schematic files and application configuration file:
	$ oam-ecs app deploy -f component1.yml,component2.yml,config.yml`,
		PreRunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			session, err := session.Default()
			if err != nil {
				return err
			}
			opts.ComponentDeployer = cloudformation.New(session)
			return nil
		}),
		RunE: runCmdE(func(cmd *cobra.Command, args []string) error {
			return opts.Execute()
		}),
	}

	cmd.Flags().StringSliceVarP(&opts.OamFiles, oamFileFlag, oamFileFlagShort, []string{}, oamFileFlagDescription)
	cmd.MarkFlagRequired(oamFileFlag)
	cmd.Flags().BoolVarP(&opts.DryRun, dryRunFlag, "", false, dryRunFlagDescription)

	return cmd
}
