// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package stack

import (
	"fmt"

	"github.com/oam-dev/oam-go-sdk/apis/core.oam.dev/v1alpha1"
)

var templateFunctions = map[string]interface{}{
	"ResolveParameterValue":       resolveOAMParameterValue,
	"ResolveTraitValue":           resolveOAMTraitValue,
	"TaskCPU":                     resolveTaskCpuValue,
	"TaskMemory":                  resolveTaskMemoryValue,
	"RequiresVolumes":             hasAnyVolumes,
	"RequiresPrivateRegistryAuth": hasAnyPullSecrets,
	"HealthCheckGracePeriod":      resolveHealthCheckGracePeriod,
}

// resolveOAMParameterValue finds the value of a named parameter
// for a given component instance configuration
func resolveOAMParameterValue(paramName string, componentConfiguration *v1alpha1.ComponentConfiguration) (string, error) {
	for _, paramValue := range componentConfiguration.ParameterValues {
		if paramValue.Name == paramName {
			return paramValue.Value, nil
		}
	}

	return "", fmt.Errorf("Could not find parameter value for name %s", paramName)
}

// resolveOAMTraitValue finds the value of a named parameter
// for a given component instance configuration
func resolveOAMTraitValue(traitName string, propertyName string, defaultValue int32, componentConfiguration *v1alpha1.ComponentConfiguration) int32 {
	if componentConfiguration.ExistTrait(traitName) {
		_, _, properties := componentConfiguration.ExtractTrait(traitName)

		if val, ok := properties[propertyName]; ok {
			return int32(val.(float64))
		}
	}

	return defaultValue
}

// hasAnyVolumes checks whether at least one of the containers requires a volume
func hasAnyVolumes(containers []v1alpha1.Container) bool {
	hasVolumes := false

	for _, container := range containers {
		if len(container.Resources.Volumes) > 0 {
			hasVolumes = true
			break
		}
	}

	return hasVolumes
}

// hasAnyPullSecrets checks whether at least one of the containers provides an image pull secret
func hasAnyPullSecrets(containers []v1alpha1.Container) bool {
	hasAnyPullSecrets := false

	for _, container := range containers {
		if container.ImagePullSecret != "" {
			hasAnyPullSecrets = true
			break
		}
	}

	return hasAnyPullSecrets
}

// resolveHealthCheckGracePeriod finds the max grace period across all containers
func resolveHealthCheckGracePeriod(containers []v1alpha1.Container) int32 {
	gracePeriod := int32(0)

	for _, container := range containers {
		if container.LivenessProbe != nil && (container.LivenessProbe.HttpGet != nil || container.LivenessProbe.TcpSocket != nil) {
			if container.LivenessProbe.InitialDelaySeconds > gracePeriod {
				gracePeriod = container.LivenessProbe.InitialDelaySeconds
			}
		}
	}

	return gracePeriod
}

type fargateTaskSize struct {
	cpuShare  float64
	memoryMiB int64
}

func getValidFargateTaskSizes() []*fargateTaskSize {
	// .25 vCPU
	taskSizes := []*fargateTaskSize{
		&fargateTaskSize{cpuShare: .25, memoryMiB: 512},
		&fargateTaskSize{cpuShare: .25, memoryMiB: 1024},
		&fargateTaskSize{cpuShare: .25, memoryMiB: 2048},
	}

	// .5 vCPU
	for i := int64(1); i <= 4; i++ {
		taskSizes = append(taskSizes, &fargateTaskSize{cpuShare: .5, memoryMiB: i * 1024})
	}

	// 1 vCPU
	for i := int64(2); i <= 8; i++ {
		taskSizes = append(taskSizes, &fargateTaskSize{cpuShare: 1, memoryMiB: i * 1024})
	}

	// 2 vCPU
	for i := int64(4); i <= 16; i++ {
		taskSizes = append(taskSizes, &fargateTaskSize{cpuShare: 2, memoryMiB: i * 1024})
	}

	// 4 vCPU
	for i := int64(8); i <= 30; i++ {
		taskSizes = append(taskSizes, &fargateTaskSize{cpuShare: 4, memoryMiB: i * 1024})
	}

	return taskSizes
}

func getNearestFargateTaskSize(containers []v1alpha1.Container) (*fargateTaskSize, error) {
	containersCpuShare := float64(0)
	containersMemoryMiB := int64(0)

	for _, container := range containers {
		containersCpuShare += float64(container.Resources.Cpu.Required.MilliValue()) / float64(1000)
		containersMemoryMiB += container.Resources.Memory.Required.MilliValue() / int64(1000) / int64(1000000)
	}

	for _, taskSize := range getValidFargateTaskSizes() {
		if containersCpuShare <= taskSize.cpuShare && containersMemoryMiB <= taskSize.memoryMiB {
			return taskSize, nil
		}
	}

	return nil, fmt.Errorf("Could not find valid Fargate task size for the given CPU and memory requirements: %f CPU shares, %d MiB memory", containersCpuShare, containersMemoryMiB)
}

// resolveTaskCpuValue finds the closest Fargate size for the containers' CPU and memory requirements,
// and returns the Fargate CPU size
func resolveTaskCpuValue(containers []v1alpha1.Container) (string, error) {
	taskSize, err := getNearestFargateTaskSize(containers)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%.2f vcpu", taskSize.cpuShare), nil
}

// resolveTaskMemoryValue finds the closest Fargate size for the containers' CPU and memory requirements,
// and returns the Fargate memory size
func resolveTaskMemoryValue(containers []v1alpha1.Container) (string, error) {
	taskSize, err := getNearestFargateTaskSize(containers)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", taskSize.memoryMiB), nil
}
