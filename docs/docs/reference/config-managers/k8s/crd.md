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

By default, the CRD we use for the Secretless Broker is under `configurations.secretless.io`.

**Note: For this plugin to work, the broker must have [ServiceAccount privileges](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
on the deployment.**

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
