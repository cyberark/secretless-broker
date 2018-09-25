# cyberark-sidecar-injector

CyberArk Broker Sidecar Injector is a [MutatingAdmissionWebhook](https://kubernetes.io/docs/admin/admission-controllers/#mutatingadmissionwebhook-beta-in-19) server which injects configurable sidecar container(s) into a pod prior to persistence of the underlying object.

  * [TL;DR;](#tl-dr-)
  * [Introduction](#introduction)
  * [Prerequisites](#prerequisites)
    + [Mandatory TLS](#mandatory-tls)
  * [Installing the Chart](#installing-the-chart)
  * [Uninstalling the Chart](#uninstalling-the-chart)
  * [Configuration](#configuration)
    + [csrEnabled=true](#csrenabledtrue)
    + [certsSecret](#certssecret)

## TL;DR;

```bash
$ helm install -f values.yaml .
```

## Introduction

This chart bootstraps a deployment of a CyberArk Broker Sidecar Injector MutatingAdmissionWebhook server including the Service and MutatingWebhookConfiguration. 

## Prerequisites

- Kubernetes 1.4+ with Beta APIs enabled

### Mandatory TLS

Supporting TLS for external webhook server is required because admission is a high security operation. As part of the installation process, we need to create a TLS certificate signed by a trusted CA (shown below is the Kubernetes CA but you can use your own) to secure the communication between the webhook server and apiserver. For the complete steps of creating and approving Certificate Signing Requests(CSR), please refer to [Managing TLS in a cluster](https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster/).

## Installing the Chart

To install the chart with the release name `my-release`, follow the instructions in the NOTES section on how to approve the CSR:

```bash
$ helm install --name my-release \
  --set caBundle="$(kubectl -n kube-system \
    get configmap \
    extension-apiserver-authentication \
    -o=jsonpath='{.data.client-ca-file}' \
  )" \
 .
```

```
...

NOTES:
## Instructions
Before you can proceed to use the sidecar-injector, there's one last step.
You will need to approve the CSR (Certificate Signing Request) made by the sidecar-injector.
This allows the sidecar-injector to communicate securely with the Kubernetes API.

### Watch initContainer logs for when CSR is created
kubectl -n injectors logs deployment/vigilant-numbat-cyberark-sidecar-injector -c init-webhook -f

### You can check and inspect the CSR
kubectl describe csr "vigilant-numbat-cyberark-sidecar-injector.injectors"

### Approve the CSR
kubectl certificate approve "vigilant-numbat-cyberark-sidecar-injector.injectors"

Now that everything is setup you can enjoy the Cyberark Sidecar Injector.
This is the general workflow:

1. Annotate your application to enable the injector and to configure the sidecar (see README.md)
2. Webhook intercepts and injects containers as needed

Enjoy.

```

The command deploys the CyberArk Broker Sidecar Injector MutatingAdmissionWebhook on the Kubernetes cluster in the default configuration. In this configuration the chart uses the cluster CA certificate bundle with a Certificate Signing Request flow to allow TLS between the webhook server and the cluster. The caBundle is required. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the CyberArk Sidecar Injector chart and their default values.

| Parameter                     | Description                                     | Default                                                    |
| -----------------------       | ---------------------------------------------   | ---------------------------------------------------------- |
| `caBundle`        | CA certificate bundle that signs the server cert used by the webhook  | `nil` (required)                                           |
| `csrEnabled`       | Generate a private key and certificate signing request towards the Kubernetes Cluster                   | `true`                        |
| `certsSecret`       | Private key and signed certificate used by the webhook server             | `nil` (required if csrEnabled is false)                |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example,

```bash
$ helm install --name my-release \
   --set csrEnabled="false" \
   --set certsSecret="some-secret" \
   --set caBundle="-----BEGIN CERTIFICATE-----..." \
   .
```

The above command creates a sidecar injector deployment, retrieves the private key and signed certificate from the `certsSecret` value and uses the `caBundle` value in the associated MutatingWebhookConfiguration. Note that `caBundle` is the certificate that signs the injector webhook server cert.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
$ helm install --name my-release -f values.yaml .
```

### certsSecret

`certsSecret` is a Kubernetes Secret containing private key and signed certificate (on paths key.pem and cert.pem, respectively)
 used by the webhook server. 

It is required for the private key and signed certificate pair to contain entries for the DNS name of the webhook service, i.e., <service name>.<namespace>.svc, or the URL of the webhook server.

### caBundle

`caBundle` is the **required** CA certificate bundle that signs the server cert used by the webhook server. It is used in the MutatingWebhookConfiguration for the release.

### csrEnabled

When `csrEnabled` is set to `true`, the chart generate a private key and certificate signing request (CSR) towards the Kubernetes Cluster, and waits until the CSR is approved before deploying the sidecar injector. 

The private key and certificate will be stored in a secret created as part of the release.

The `caBundle` in this case is the Kubernetes cluster CA certificate. This can be retrieve as follows:

```
kubectl -n kube-system \
  get configmap \
  extension-apiserver-authentication \
  -o=jsonpath='{.data.client-ca-file}'
```
