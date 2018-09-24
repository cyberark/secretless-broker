---
title: Configuration Managers
id: configfile
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/config-managers/k8s/crd.html
---

## Kubernetes Custom Resource Definition (CRD)

The Kubernetes CRD plugin (`k8s/crd`) allows the use of Kubernetes-specific
custom resource definitions to trigger and specify the configuration for Secretless Broker.

### Operation

When this configuration manager plugin starts, a new CRD is registered (if not already available)
as `configurations.secretless.io` after which the configuration ID specified by the name after the `#`
symbol on the CLI will be watched in that CRD namespace. As soon as the configuration with that ID is
defined or updated in Kubernetes, Secretless will notify the main daemon to load the new configuration
and soft-reload itself. Any number of Secretless Broker instances can watch the same configuration ID.

## Kubernetes API Permissions

**Note: For this plugin to work, the broker must have [ServiceAccount privileges](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
on the deployment.**

The basic role configuration which allows Secretless Broker to work within a Kubernetes cluster without full cluster administrator permissions is below:
```
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: secretless-crd
rules:
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - customresourcedefinitions
  verbs:
  - create
  - get
  - watch
  - list
- apiGroups: [""]
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - secretless.io
  resources:
  - configurations
  verbs:
  - get
  - list
  - watch

---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: secretless-crd

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: secretless-crd
subjects:
- kind: ServiceAccount
  name: secretless-crd
  namespace: default
roleRef:
  kind: ClusterRole
  name: secretless-crd
  apiGroup: rbac.authorization.k8s.io
```

After defining the `ServiceAccount`, `ClusterRole`, and `ClusterRoleBinding`, you can then use it in your deployment with a `serviceAccountName` parameter:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: secretless-k8s-test
spec:
  ...
  template:
    ...
    spec:
      serviceAccountName: secretless-crd
      containers:
      ...
```

## Examples

Start broker and watch for `secretless-example-config` resource in `configurations.secretless.io` resource
namespace:
```
$ secretless-broker -config-mgr k8s/crd#secretless-example-config
```

Any additions or updates of `secretless-example-config` resource in `configurations.secretless.io` CRD
namespace will trigger a reload of the broker:

Note: You can find `sbconfig-example.yaml` and other referenced configuration files in the `resource-definitions`
directory of the code repository.
```
$ # This command should trigger a reload of the broker from earlier with the configuration specified in
$ # the file
$ kubectl apply -f resource-definitions/sbconfig-example.yaml
```
