//go:generate packr2

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package templates

import "github.com/gobuffalo/packr/v2"

// Box can be used to read in templates from the templates directory.
// For example, templates.Box().FindString("core.oam.dev/v1alpha1.Server/cf.yml").
func Box() *packr.Box {
	return packr.New("templates", "./")
}
