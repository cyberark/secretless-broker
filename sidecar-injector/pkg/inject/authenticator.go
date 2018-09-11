package inject

import (
    corev1 "k8s.io/api/core/v1"
)

type AuthenticatorSidecarConfig struct {
    connectionConfigMap string
    containerMode string
    containerName string
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
    return &PatchConfig{
        Containers: []corev1.Container{
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
                    envVarFromConfigMap("CONJUR_VERSION", authConfig.connectionConfigMap),
                    envVarFromConfigMap("CONJUR_APPLIANCE_URL", authConfig.connectionConfigMap),
                    envVarFromConfigMap("CONJUR_AUTHN_URL", authConfig.connectionConfigMap),
                    envVarFromConfigMap("CONJUR_ACCOUNT", authConfig.connectionConfigMap),
                    envVarFromConfigMap("CONJUR_SSL_CERTIFICATE", authConfig.connectionConfigMap),
                    envVarFromConfigMap("CONJUR_AUTHN_LOGIN", authConfig.connectionConfigMap),
                },
                VolumeMounts: []corev1.VolumeMount{
                    {
                        Name:      "conjur-access-token",
                        MountPath: "/run/conjur",
                    },
                },
            },
        },
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
