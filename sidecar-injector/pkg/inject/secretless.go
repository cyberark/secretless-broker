package inject

import (
    corev1 "k8s.io/api/core/v1"
)

// generateSecretlessSidecarConfig generates PatchConfig from a given secretlessConfigMapName
func generateSecretlessSidecarConfig(secretlessConfigMapName string) *PatchConfig {
    return &PatchConfig{
        Containers: []corev1.Container{
            {
                Name:            "secretless",
                Image:           "cyberark/secretless-broker:latest",
                Args:            []string{"-f", "/etc/secretless/secretless.yml"},
                ImagePullPolicy: "Always",
                VolumeMounts: []corev1.VolumeMount{
                    {
                        Name:      "secretless-config",
                        ReadOnly:  true,
                        MountPath: "/etc/secretless",
                    },
                },
            },
        },
        Volumes: []corev1.Volume{
            {
                Name: "secretless-config",
                VolumeSource: corev1.VolumeSource{
                    ConfigMap: &corev1.ConfigMapVolumeSource{
                        LocalObjectReference: corev1.LocalObjectReference{
                            Name: secretlessConfigMapName,
                        },
                    },
                },
            },
        },
    }
}
