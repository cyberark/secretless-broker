package inject

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/golang/glog"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubernetes/pkg/apis/core/v1"
	"k8s.io/api/admission/v1beta1"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)


func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
}

// applyDefaultsWorkaround applies a defaulting on Container and Volume specs to address this issue (https://github.com/kubernetes/kubernetes/issues/57982)
func applyDefaultsWorkaround(containers []corev1.Container, volumes []corev1.Volume) {
	defaulter.Default(&corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: containers,
			Volumes:    volumes,
		},
	})
}

var ignoredNamespaces = []string{
	metav1.NamespaceSystem,
	metav1.NamespacePublic,
}

type WebhookServer struct {
	Server *http.Server
}

// Webhook Server parameters
type WebhookServerParameters struct {
	Port     int    // webhook Server port
	CertFile string // path to the x509 certificate for https
	KeyFile  string // path to the x509 private key matching `CertFile`
}

func failWithResponse(errMsg string) *v1beta1.AdmissionResponse {
    glog.Infof(errMsg)
    return &v1beta1.AdmissionResponse{
        Result: &metav1.Status{
            Message: errMsg,
        },
    }
}

// main mutation process
func (whsvr *WebhookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		return failWithResponse(fmt.Sprintf("Could not unmarshal raw object: %v", err))
	}

	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v rfc6902PatchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)

	// determine whether to perform mutation
	if !mutationRequired(ignoredNamespaces, &pod.ObjectMeta) {
		glog.Infof("Skipping mutation for %s/%s due to policy check", req.Namespace, pod.Name)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

    injectType, err := getAnnotation(&pod.ObjectMeta, annotationInjectTypeKey)
    containerMode, err := getAnnotation(&pod.ObjectMeta, annotationContainerModeKey)
    containerName, err := getAnnotation(&pod.ObjectMeta, annotationContainerNameKey)


    var sidecarConfig *PatchConfig
    switch injectType {
    case "secretless":
        secretlessConfigMapName, err := getAnnotation(&pod.ObjectMeta, annotationSecretlessConfigKey)
        if err != nil {
            return failWithResponse(fmt.Sprintf("Mutation failed for pod %s, in namespace %s, due to %s", pod.Name, req.Namespace, err.Error()))
        }

        conjurConnConfigMapName, _ := getAnnotation(&pod.ObjectMeta, annotationConjurConnConfigKey)
        conjurAuthConfigMapName, _ := getAnnotation(&pod.ObjectMeta, annotationConjurAuthConfigKey)

        sidecarConfig = generateSecretlessSidecarConfig(
            secretlessConfigMapName,
            conjurConnConfigMapName,
            conjurAuthConfigMapName)
        break;
    case "authenticator":
        conjurAuthConfigMapName, err := getAnnotation(&pod.ObjectMeta, annotationConjurAuthConfigKey)
        if err != nil {
            return failWithResponse(fmt.Sprintf("Mutation failed for pod %s, in namespace %s, due to %s", pod.Name, req.Namespace, err.Error()))
        }

        conjurConnConfigMapName, err := getAnnotation(&pod.ObjectMeta, annotationConjurConnConfigKey)
        if err != nil {
            return failWithResponse(fmt.Sprintf("Mutation failed for pod %s, in namespace %s, due to %s", pod.Name, req.Namespace, err.Error()))
        }

        switch containerMode {
        case "sidecar", "init", "":
            break;
        default:
            return failWithResponse(fmt.Sprintf("Mutation failed for pod %s, in namespace %s, due to %s value (%s) not supported", pod.Name, req.Namespace, annotationContainerModeKey, containerMode))
        }

        sidecarConfig = generateAuthenticatorSidecarConfig(AuthenticatorSidecarConfig{
            conjurConnConfigMapName: conjurConnConfigMapName,
            conjurAuthConfigMapName: conjurAuthConfigMapName,
            containerMode:           containerMode,
            containerName:           containerName,
        })
        break;
    default:
        errMsg := fmt.Sprintf("Mutation failed for pod %s, in namespace %s, due to invalid inject type annotation value = %s", pod.Name, req.Namespace, injectType)
        glog.Infof(errMsg)
        return &v1beta1.AdmissionResponse{
            Result: &metav1.Status{
                Message: errMsg,
            },
        }
    }

	// Workaround: https://github.com/kubernetes/kubernetes/issues/57982
	applyDefaultsWorkaround(sidecarConfig.Containers, sidecarConfig.Volumes)
	annotations := map[string]string{annotationStatusKey: "injected"}
	patchBytes, err := createPatch(&pod, sidecarConfig, annotations)
	if err != nil {
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	glog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// Serve method for webhook Server
func (whsvr *WebhookServer) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expecting application/json", contentType)
		http.Error(w, "invalid Content-Type, expecting `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("could not decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		admissionResponse = whsvr.mutate(&ar)
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		glog.Errorf("could not encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.Infof("Ready to write response ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("could not write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
