package inject

import (
    corev1 "k8s.io/api/core/v1"
)

type AuthenticatorSidecarConfig struct {
    conjurConnConfigMapName string
    conjurAuthConfigMapName string
    containerMode           string
    containerName           string
}

func (authConfig AuthenticatorSidecarConfig) ContainerNameOrDefault() string {
    name := "authenticator"
    if authConfig.containerName != "" {
        name = authConfig.containerName
    }

    return name
}

// generateAuthenticatorSidecarConfig generates PatchConfig from a given AuthenticatorSidecarConfig
func generateAuthenticatorSidecarConfig(authConfig AuthenticatorSidecarConfig) *PatchConfig {
    var containers, initContainers []corev1.Container

    candidates := []corev1.Container{
        {
            Name:            authConfig.ContainerNameOrDefault() ,
            Image:           "cyberark/conjur-kubernetes-authenticator:latest",
            ImagePullPolicy: "IfNotPresent",
            Env: []corev1.EnvVar{
                envVarFromFieldPath("MY_POD_NAME", "metadata.name"),
                envVarFromFieldPath("MY_POD_NAMESPACE", "metadata.namespace"),
                envVarFromFieldPath("MY_POD_IP", "status.podIP"),
                {
                    Name: "CONJUR_AUTHN_TOKEN_FILE",
                    Value: "/run/conjur/conjur-access-token",
                },
                {
                    Name: "CONTAINER_MODE",
                    Value: authConfig.containerMode,
                },
                envVarFromConfigMap("CONJUR_VERSION", authConfig.conjurConnConfigMapName),
                envVarFromConfigMap("CONJUR_APPLIANCE_URL", authConfig.conjurConnConfigMapName),
                envVarFromConfigMap("CONJUR_AUTHN_URL", authConfig.conjurConnConfigMapName),
                envVarFromConfigMap("CONJUR_ACCOUNT", authConfig.conjurConnConfigMapName),
                envVarFromConfigMap("CONJUR_SSL_CERTIFICATE", authConfig.conjurConnConfigMapName),
                envVarFromConfigMap("CONJUR_AUTHN_LOGIN", authConfig.conjurAuthConfigMapName),
            },
            VolumeMounts: []corev1.VolumeMount{
                {
                    Name:      "conjur-access-token",
                    MountPath: "/run/conjur",
                },
            },
        },
    }

    if authConfig.containerMode == "init" {
        initContainers = candidates
    } else {
        containers = candidates
    }

    return &PatchConfig{
        Containers: containers,
        InitContainers: initContainers,
        Volumes: []corev1.Volume{
            {
                Name: "conjur-access-token",
                VolumeSource: corev1.VolumeSource{
                    EmptyDir: &corev1.EmptyDirVolumeSource{
                        Medium:    "Memory",
                    },
                },
            },
        },
    }
}
