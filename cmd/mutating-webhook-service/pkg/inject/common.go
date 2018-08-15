package inject

import (
	"strings"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Check whether the target resoured need to be mutated
func mutationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernete system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip mutation for %v for it' in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	status := annotations[admissionWebhookAnnotationStatusKey]

	// determine whether to perform mutation based on annotation for the target resource
	var required bool
	if strings.ToLower(status) == "injected" {
		required = false;
	} else {
		switch strings.ToLower(annotations[admissionWebhookAnnotationInjectKey]) {
		default:
			required = false
		case "y", "yes", "true", "on":
			required = true
		}
	}

	glog.Infof("Mutation policy for %v/%v: status: %q required:%v", metadata.Namespace, metadata.Name, status, required)
	return required
}

// generateSidecarConfig generates Config from a given secretlessConfigMapName
func generateSidecarConfig(secretlessConfigMapName string) *Config {
	return &Config{
		Containers: []corev1.Container{
			{
				Name:            "secretless",
				Image:           "cyberark/secretless-broker:latest",
				Args:            []string{"-f", "/etc/secretless/secretless.yml"},
				ImagePullPolicy: "IfNotPresent",
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

// getSecretlessConfigMapName attempts to find the string value of
// the admissionWebhookAnnotationConfigKey annotation inside a given ObjectMeta.
// It returns a tuple (string, bool):
// the string value (or empty) and a boolean of whether the annotation was present
func getSecretlessConfigMapName(metadata *metav1.ObjectMeta) (string, bool) {
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	value, hasKey := annotations[admissionWebhookAnnotationConfigKey]

	return value, hasKey
}
