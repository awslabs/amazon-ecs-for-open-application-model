// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package types

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/olekukonko/tablewriter"
)

// EnvironmentInput holds the fields required to interact with an environment.
type EnvironmentInput struct {
}

// Environment represents the configuration of a particular environment
type Environment struct {
	StackName    string
	StackOutputs map[string]string
}

// CreateEnvironmentResponse holds the created environment on successful deployment.
// Otherwise, the environment is set to nil and a descriptive error is returned.
type CreateEnvironmentResponse struct {
	Env *Environment
	Err error
}

func (env *Environment) Display() {
	fmt.Printf("\nEnvironment: %s\n\n", env.StackName)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Environment Attribute", "Value"})
	table.SetBorder(false)

	keys := make([]string, 0, len(env.StackOutputs))
	for key := range env.StackOutputs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		formattedKey := strings.Title(strings.ToLower(strcase.ToDelimited(key, ' ')))
		formattedKey = strings.ReplaceAll(formattedKey, "Cloud Formation", "CloudFormation")
		formattedKey = strings.ReplaceAll(formattedKey, "Ecs", "ECS")
		table.Append([]string{formattedKey, env.StackOutputs[key]})
	}

	table.Render()
	fmt.Println("")
}
