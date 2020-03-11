// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"fmt"
	"strings"

	deploy "github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/deploy/cloudformation"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/color"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/term/progress"
)

// ResourceMatcher is a function that returns true if the resource event matches a criteria.
type ResourceMatcher func(deploy.Resource) bool

// HumanizeResourceEvents groups raw deploy events under human-friendly tab-separated texts
// that can be passed into the Events() method. Every text to display starts with status in progress.
// For every resource event that belongs to a text, we  preserve failure events if there was one.
// Otherwise, the text remains in progress until the expected number of resources reach the complete status.
func HumanizeResourceEvents(orderedTexts []progress.Text, resourceEvents []deploy.ResourceEvent, matcher map[progress.Text]ResourceMatcher, wantedCount map[progress.Text]int) []progress.TabRow {
	// Assign a status to text from all matched events.
	statuses := make(map[progress.Text]progress.Status)
	reasons := make(map[progress.Text]string)
	for text, matches := range matcher {
		statuses[text] = progress.StatusInProgress
		for _, resourceEvent := range resourceEvents {
			if !matches(resourceEvent.Resource) {
				continue
			}
			if oldStatus, ok := statuses[text]; ok && oldStatus == progress.StatusFailed {
				// There was a failure event, keep its status.
				continue
			}
			status := toStatus(resourceEvent.Status)
			if status == progress.StatusComplete || status == progress.StatusSkipped {
				// If there are more resources that needs to have StatusComplete then the text should remain in StatusInProgress.
				wantedCount[text] = wantedCount[text] - 1
				if wantedCount[text] > 0 {
					status = progress.StatusInProgress
				}
			}
			statuses[text] = status
			reasons[text] = resourceEvent.StatusReason
		}
	}

	// Serialize the text and status to a format digestible by Events().
	var rows []progress.TabRow
	for _, text := range orderedTexts {
		status, ok := statuses[text]
		if !ok {
			continue
		}
		coloredStatus := fmt.Sprintf("[%s]", status)
		if status == progress.StatusInProgress {
			coloredStatus = color.Grey.Sprint(coloredStatus)
		}
		if status == progress.StatusFailed {
			coloredStatus = color.Red.Sprint(coloredStatus)
		}

		rows = append(rows, progress.TabRow(fmt.Sprintf("%s\t%s", color.Grey.Sprint(text), coloredStatus)))
		if status == progress.StatusFailed {
			rows = append(rows, progress.TabRow(fmt.Sprintf("  %s\t", reasons[text])))
		}
	}
	return rows
}

func toStatus(s string) progress.Status {
	if strings.HasSuffix(s, "FAILED") {
		return progress.StatusFailed
	}
	if strings.HasSuffix(s, "COMPLETE") {
		return progress.StatusComplete
	}
	if strings.HasSuffix(s, "SKIPPED") {
		return progress.StatusSkipped
	}
	return progress.StatusInProgress
}
