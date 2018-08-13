package kubernetes;

import (
	"io"
	"bufio"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/api/core/v1"
	"k8s.io/api/apps/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

const (
	ContainerName = "secretless"
	ContainerImage = "cyberark/secretless"
	AnnotationConfig = "secretlessConfig"
	VolumeConfigName = "secretless-config"
	VolumeConjurName = "secretless-conjur-access-token"
	VolumeConfigMountPath = "/etc/secretless"
	VolumeConjurMountPath = "/var/run/conjur"
)

func addVolume(podSpec *v1.PodSpec) {
	secretlessVolume := v1.Volume{
		Name: VolumeConjurName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium: v1.StorageMediumMemory,
			},
		},
	}

	podSpec.Volumes = append(podSpec.Volumes, secretlessVolume)
}

func addVolumeMount(container *v1.Container) {
	secretlessVolumeMount := v1.VolumeMount{
		Name: VolumeConjurName,
		ReadOnly: true,
		MountPath: VolumeConjurMountPath,
	}

	container.VolumeMounts = append(container.VolumeMounts, secretlessVolumeMount)
}

func addContainer(podSpec *v1.PodSpec, configName string) {
	for _, container := range podSpec.Containers {
		addVolumeMount(&container)
	}

	secretlessContainer := v1.Container {
		Name: ContainerName,
		Image: ContainerImage,
		VolumeMounts: []v1.VolumeMount{
			v1.VolumeMount{
				Name: VolumeConjurName,
				MountPath: VolumeConjurMountPath,
			},
			v1.VolumeMount{
				Name: configName,
				ReadOnly: true,
				MountPath: VolumeConfigMountPath,
			},
		},
	}

	podSpec.Containers = append(podSpec.Containers, secretlessContainer)
}

func addConfigMap(podSpec *v1.PodSpec, configName string) {
	configMap := v1.Volume{
		Name: VolumeConfigName,
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource {
				LocalObjectReference: v1.LocalObjectReference{
					Name: configName,
				},
			},
		},
	}

	podSpec.Volumes = append(podSpec.Volumes, configMap)
}

func configReference(objectMeta metav1.ObjectMeta) string {
	return objectMeta.Annotations[AnnotationConfig]
}

func inject(obj runtime.Object) {
	var podSpec *v1.PodSpec
	var config string
	switch o := obj.(type) {
	case *v1.Pod:
		config = configReference(o.ObjectMeta)
		if config != "" {
			podSpec = &o.Spec
		}
	case *v1beta1.Deployment:
		config = configReference(o.ObjectMeta)
		if config != "" { 
			podSpec = &o.Spec.Template.Spec 
		}
	case *appsv1.ReplicaSet:
		config = configReference(o.ObjectMeta)
		if config != "" { 
			podSpec = &o.Spec.Template.Spec 
		}
	default:
		// not applicable
	}

	if podSpec != nil {
		addContainer(podSpec, config)
		addVolume(podSpec)
	}
}

// InjectManifest injects the Secretless sidecar (along with volumes and mounts)
// into a Pod, Deployment or ReplicaSet. The object targeted for injection must
// be annotated with the key `AnnotationConfig` and the value referencing a
// ConfigMap to be used as the Secretless configuration.
func InjectManifest(manifest io.Reader, out io.Writer) error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	reader := yaml.NewYAMLReader(bufio.NewReader(manifest))
	writer := json.YAMLFramer.NewFrameWriter(out)
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

	for {
		doc, err := reader.Read()
		if err != nil {
			break
		}

		obj, _, err := decode([]byte(doc), nil, nil)
		if err != nil {
			return err
		}

		inject(obj)
		err = serializer.Encode(obj, writer)
		if err != nil {
			return err
		}
	}

	return nil
}