// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"testing"

	deploy "github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/progress"
	"github.com/stretchr/testify/require"
)

func TestHumanizeResourceEvents(t *testing.T) {
	testCases := map[string]struct {
		inResourceEvents []deploy.ResourceEvent
		inDisplayOrder   []progress.Text
		inMatcher        map[progress.Text]ResourceMatcher

		wantedEvents []progress.TabRow
	}{
		"grabs the first failure": {
			inResourceEvents: []deploy.ResourceEvent{
				{
					Resource: deploy.Resource{
						LogicalName: "VPC",
						Type:        "AWS::EC2::VPC",
					},
					Status:       "CREATE_FAILED",
					StatusReason: "first failure",
				},
				{
					Resource: deploy.Resource{
						LogicalName: "VPC",
						Type:        "AWS::EC2::VPC",
					},
					Status:       "CREATE_FAILED",
					StatusReason: "second failure",
				},
			},
			inDisplayOrder: []progress.Text{"vpc"},
			inMatcher: map[progress.Text]ResourceMatcher{
				"vpc": func(resource deploy.Resource) bool {
					return resource.Type == "AWS::EC2::VPC"
				},
			},

			wantedEvents: []progress.TabRow{"vpc\t[Failed]", "  first failure\t"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := HumanizeResourceEvents(tc.inDisplayOrder, tc.inResourceEvents, tc.inMatcher, nil)

			require.Equal(t, tc.wantedEvents, got)
		})
	}
}
