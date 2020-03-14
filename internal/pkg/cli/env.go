// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
)

// BuildEnvCmd is the top level command for environments
func BuildEnvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Environment commands",
		Long:  `Commands for working with the oam-ecs environment.`,
	}

	cmd.AddCommand(BuildDeployEnvironmentCmd())
	cmd.AddCommand(BuildShowEnvironmentCmd())
	cmd.AddCommand(BuildDeleteEnvironmentCmd())

	return cmd
}
