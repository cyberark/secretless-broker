package inject

import (
	corev1 "k8s.io/api/core/v1"
)

type PatchConfig struct {
	InitContainers []corev1.Container `yaml:"initContainers"`
	Containers []corev1.Container `yaml:"containers"`
	Volumes    []corev1.Volume    `yaml:"volumes"`
}
