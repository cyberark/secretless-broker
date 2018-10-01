package inject

import (
    corev1 "k8s.io/api/core/v1"
)

// generateSecretlessSidecarConfig generates PatchConfig from a given secretlessConfigMapName
func generateSecretlessSidecarConfig(secretlessConfigMapName, conjurConnConfigMapName, conjurAuthConfigMapName string) *PatchConfig {
    envvars := []corev1.EnvVar{
        envVarFromFieldPath("MY_POD_NAME", "metadata.name"),
        envVarFromFieldPath("MY_POD_NAMESPACE", "metadata.namespace"),
        envVarFromFieldPath("MY_POD_IP", "status.podIP"),
    }

    if conjurConnConfigMapName != "" || conjurAuthConfigMapName != "" {
        envvars = append(envvars,
            envVarFromConfigMap("CONJUR_VERSION", conjurConnConfigMapName),
            envVarFromConfigMap("CONJUR_APPLIANCE_URL", conjurConnConfigMapName),
            envVarFromConfigMap("CONJUR_AUTHN_URL", conjurConnConfigMapName),
            envVarFromConfigMap("CONJUR_ACCOUNT", conjurConnConfigMapName),
            envVarFromConfigMap("CONJUR_SSL_CERTIFICATE", conjurConnConfigMapName),
            envVarFromConfigMap("CONJUR_AUTHN_LOGIN", conjurAuthConfigMapName))
    }

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
                Env: envvars,
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
