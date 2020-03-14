// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package cli

// Long flag names.
const (
	oamFileFlag = "filename"
	dryRunFlag  = "dry-run"
)

// Short flag names.
// A short flag only exists if the flag is mandatory by the command.
const (
	oamFileFlagShort = "f"
)

// Descriptions for flags.
const (
	oamFileFlagDescription       = "Path to a file containing OAM component schematics or OAM application configuration. Multiple files can be provided either by repeating the flag for each file, or with a comma-delimited list of files."
	dryRunFlagDescription        = "Write out an infrastructure template to a file instead of deploying the infrastructure"
	appConfigFileFlagDescription = "Path to a file containing an OAM application configuration."
)
