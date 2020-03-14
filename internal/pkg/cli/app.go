// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"
)

// BuildAppCmd is the top level command for applications
func BuildAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Application commands",
		Long:  `Commands for working with an oam-ecs application.`,
	}

	cmd.AddCommand(BuildDeployAppCmd())

	return cmd
}
