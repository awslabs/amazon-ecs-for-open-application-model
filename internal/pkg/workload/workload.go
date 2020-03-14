// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package workload defines OAM workload descriptions
package workload

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/log"
	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	workerComponentWorkloadType = "core.oam.dev/v1alpha1.Worker"
	serverComponentWorkloadType = "core.oam.dev/v1alpha1.Server"
)

type OamWorkloadProps struct {
	OamFiles []string
}

type OamWorkload struct {
	ApplicationConfiguration *v1alpha1.ApplicationConfiguration
	ComponentSchematics      map[string]*v1alpha1.ComponentSchematic
}

func NewOamWorkload(input *OamWorkloadProps) (*OamWorkload, error) {
	var applicationConfiguration *v1alpha1.ApplicationConfiguration
	componentSchematics := make(map[string]*v1alpha1.ComponentSchematic)

	v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
	decode := scheme.Codecs.UniversalDeserializer().Decode

	// Parse all of the app config and component schematics from the given files
	for _, fileLocation := range input.OamFiles {
		fileContents, err := ioutil.ReadFile(fileLocation)
		if err != nil {
			log.Errorf("Failed to read file %s\n", fileLocation)
			return nil, err
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
				return nil, err
			}
			chunk = chunk[:n]

			obj, kind, err := decode(chunk, nil, nil)
			if err != nil {
				log.Errorf("Failed to parse file %s\n", fileLocation)
				return nil, err
			}

			switch obj.(type) {
			case *v1alpha1.ApplicationConfiguration:
				if applicationConfiguration != nil {
					log.Errorf("File %s contains an ApplicationConfiguration, but one has already been found\n", fileLocation)
					return nil, fmt.Errorf("Multiple application configuration files found, only one is allowed per application")
				}
				applicationConfiguration = obj.(*v1alpha1.ApplicationConfiguration)
			case *v1alpha1.ComponentSchematic:
				schematic := obj.(*v1alpha1.ComponentSchematic)

				if schematic.Spec.WorkloadType != workerComponentWorkloadType &&
					schematic.Spec.WorkloadType != serverComponentWorkloadType {
					log.Errorf("Component schematic %s is an invalid workload type\n", schematic.Name)
					return nil, fmt.Errorf("Workload type is %s, only %s and %s are supported",
						schematic.Spec.WorkloadType,
						workerComponentWorkloadType,
						serverComponentWorkloadType)
				}

				componentSchematics[schematic.Name] = schematic
			default:
				log.Errorf("Found invalid object in file %s\n", fileLocation)
				return nil, fmt.Errorf("Object type %s is not supported", kind)
			}
			log.Successf("Read %s from file %s\n", kind, fileLocation)
		}
	}

	if applicationConfiguration == nil {
		log.Errorf("No application configuration found in given files %s\n", strings.Join(input.OamFiles, ", "))
		return nil, fmt.Errorf("Application configuration is required")
	}

	return &OamWorkload{
		ApplicationConfiguration: applicationConfiguration,
		ComponentSchematics:      componentSchematics,
	}, nil
}
