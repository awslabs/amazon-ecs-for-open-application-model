// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cli contains the oam-ecs subcommands.
package cli

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/aws/session"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation/types"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/log"
	termprogress "github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/progress"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	workerComponentWorkloadType = "core.oam.dev/v1alpha1.Worker"
	serverComponentWorkloadType = "core.oam.dev/v1alpha1.Server"

	environmentName = "oam-ecs"

	dryRunComponentStart     = "Computing infrastructure template for the component instance %s."
	dryRunComponentFailed    = "Failed to generate infrastructure template for the component instance %s."
	dryRunComponentSucceeded = "Wrote infrastructure template to disk for component instance %s: %s"
	deployComponentStart     = "Deploying infrastructure changes for the component instance %s."
	deployComponentFailed    = "Failed to deploy infrastructure changes for the component instance %s."
	deployComponentSucceeded = "Deployed component instance %s in CloudFormation stack %s.\n"
)

type cfComponentDeployer interface {
	DeployComponent(component *types.DeployComponentInput) (*types.Component, error)
	DryRunComponent(component *types.DeployComponentInput) (string, error)
}

// ApplyOpts holds the configuration needed to provision an application.
type ApplyOpts struct {
	// Fields with matching flags
	OamFiles []string
	DryRun   bool

	prog              progress
	ComponentDeployer cfComponentDeployer
}

// NewApplyOpts initiates the fields to provision an application.
func NewApplyOpts() *ApplyOpts {
	return &ApplyOpts{
		prog: termprogress.NewSpinner(),
	}
}

func (opts *ApplyOpts) parseOamFiles(oamFiles []string) (*v1alpha1.ApplicationConfiguration, map[string]*v1alpha1.ComponentSchematic, error) {
	var applicationConfiguration *v1alpha1.ApplicationConfiguration
	componentSchematics := make(map[string]*v1alpha1.ComponentSchematic)

	v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode

	// Parse all of the app config and component schematics from the given files
	for _, fileLocation := range oamFiles {
		fileContents, err := ioutil.ReadFile(fileLocation)
		if err != nil {
			log.Errorf("Failed to read file %s\n", fileLocation)
			return nil, nil, err
		}

		// Split the file into potentially multiple YAML documents delimited by '\n---'
		reader := yaml.NewDocumentDecoder(ioutil.NopCloser(strings.NewReader(string(fileContents))))
		for {
			chunk := make([]byte, len(fileContents))
			n, err := reader.Read(chunk)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("Failed to read file %s\n", fileLocation)
				return nil, nil, err
			}
			chunk = chunk[:n]

			obj, kind, err := decode(chunk, nil, nil)
			if err != nil {
				log.Errorf("Failed to parse file %s\n", fileLocation)
				return nil, nil, err
			}

			switch obj.(type) {
			case *v1alpha1.ApplicationConfiguration:
				if applicationConfiguration != nil {
					log.Errorf("File %s contains an ApplicationConfiguration, but one has already been found\n", fileLocation)
					return nil, nil, fmt.Errorf("Multiple application configuration files found, only one is allowed per application")
				}
				applicationConfiguration = obj.(*v1alpha1.ApplicationConfiguration)
			case *v1alpha1.ComponentSchematic:
				schematic := obj.(*v1alpha1.ComponentSchematic)

				if schematic.Spec.WorkloadType != workerComponentWorkloadType && schematic.Spec.WorkloadType != serverComponentWorkloadType {
					log.Errorf("Component schematic %s is an invalid workload type\n", schematic.Name)
					return nil, nil, fmt.Errorf("Workload type is %s, only %s and %s are supported", schematic.Spec.WorkloadType, workerComponentWorkloadType, serverComponentWorkloadType)
				}

				componentSchematics[schematic.Name] = schematic
			default:
				log.Errorf("Found invalid object in file %s\n", fileLocation)
				return nil, nil, fmt.Errorf("Object type %s is not supported", kind)
			}
			log.Successf("Read %s from file %s\n", kind, fileLocation)
		}
	}

	if applicationConfiguration == nil {
		log.Errorf("No application configuration found in given files %s\n", strings.Join(oamFiles, ", "))
		return nil, nil, fmt.Errorf("Application configuration is required")
	}

	// Validate app config and component schematics
	for _, component := range applicationConfiguration.Spec.Components {
		_, ok := componentSchematics[component.ComponentName]
		if !ok {
			log.Errorf("Could not find component schematic for %s\n", component.ComponentName)
			return nil, nil, fmt.Errorf("Application configuration refers to component %s, but no file provided the component schematic", component.ComponentName)
		}
	}

	return applicationConfiguration, componentSchematics, nil
}

func (opts *ApplyOpts) newDeployComponentInput(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration, schematic *v1alpha1.ComponentSchematic) (*types.DeployComponentInput, error) {
	// TODO validate that following are not set: osType, arch, volume disk, volume sharing policy,
	// 				container extended resource, container config file, container readiness probe,
	//				container liveness probe failure threshold/httpGet/tcpSocket

	ecsSettings := &types.ECSWorkloadSettings{}

	environment := &types.ComponentEnvironment{
		Name: environmentName,
	}

	return &types.DeployComponentInput{
		ApplicationConfiguration: application,
		ComponentConfiguration:   componentInstance,
		Component:                schematic,
		WorkloadSettings:         ecsSettings,
		Environment:              environment,
	}, nil
}

func (opts *ApplyOpts) dryRunComponentInstance(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration, schematic *v1alpha1.ComponentSchematic) error {
	deployComponentInput, err := opts.newDeployComponentInput(application, componentInstance, schematic)
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

func (opts *ApplyOpts) deployComponentInstance(application *v1alpha1.ApplicationConfiguration, componentInstance *v1alpha1.ComponentConfiguration, schematic *v1alpha1.ComponentSchematic) error {
	deployComponentInput, err := opts.newDeployComponentInput(application, componentInstance, schematic)
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

	for key, value := range component.StackOutputs {
		log.Infof("\t%s:\n\t\t%s\n\n", key, value)
	}

	return nil
}

// Execute parses the OAM files, translates them into infrastructure definitions, and deploys the infrastructure
func (opts *ApplyOpts) Execute() error {
	applicationConfiguration, componentSchematics, err := opts.parseOamFiles(opts.OamFiles)
	if err != nil {
		return err
	}

	for _, componentInstance := range applicationConfiguration.Spec.Components {
		schematic, _ := componentSchematics[componentInstance.ComponentName]
		if opts.DryRun {
			err = opts.dryRunComponentInstance(applicationConfiguration, &componentInstance, schematic)
		} else {
			err = opts.deployComponentInstance(applicationConfiguration, &componentInstance, schematic)
		}

		if err != nil {
			break
		}
	}

	return err
}

// BuildApplyCmd build the command for deploying an application.
func BuildApplyCmd() *cobra.Command {
	opts := NewApplyOpts()
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Deploy the application",
		Long:  `Provisions (or updates) the Amazon ECS infrastructure for the application defined using the Open Application Model spec. All component schematics and the application configuration file for the application must be provided every time the apply command runs (this CLI does not save any state).`,
		Example: `
  Apply the application's OAM component schematic files and application configuration file:
	$ oam-ecs apply -f component1.yml,component2.yml,config.yml`,
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
