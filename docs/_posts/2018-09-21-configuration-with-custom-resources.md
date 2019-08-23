---
layout: post
title: "Using Kubernetes Custom Resources to Configure Secretless"
date: 2018-09-21 09:00:00 -0600
author: Geri Jennings
categories: blog
published: true
image: secretless_logo_blog.jpg
thumb: secretless_logo_blog.jpg
image-alt: Secretless logo
excerpt: "Secretless Broker by default expects its configuration via file, but users
  deploying their apps to Kubernetes can also take advantage of our Secretless config
  Custom Resource to provide their Secretless Broker with its configuration"
---

# The Default Method - Configuration By File

The default method for configuring your Secretless Broker is to provide it with
a `secretless.yml` file that specifies the Service Connectors that Secretless
Broker should be running. Your `secretless.yml` might look something like:

```
version: "2"
services:
  my_webapp_connector:
    connector: basic_auth
    listenOn: tcp://0.0.0.0:8080
    credentials:
      username:
        from: env
        get: WEBAPP_USERNAME
      password:
        from: env
        get: WEBAPP_PASSWORD
    config:
      authenticateURLsMatching:
        - ^http.*
```
This example `secretless.yml` is for a webservice my app needs to connect to that
uses basic auth as its authentication scheme.

To provide this configuration to Secretless Broker when you deploy it to Kubernetes
in the same pod as your application, you could add a ConfigMap:

```
kubectl --namespace my-app-ns \
  create configmap \
  my-app-secretless-config \
  --from-file=secretless.yml
```

Then, when you deploy your app to Kubernetes, you reference the ConfigMap in your
deployment manifest:

```
containers:
  - name: my-app
    # deployment details for my-app
  - name: secretless-broker
    image: cyberark/secretless-broker:latest
    imagePullPolicy: IfNotPresent
    volumeMounts:
      - name: config
        mountPath: /etc/
        readOnly: true
volumes:
  - name: config
    configMap:
      name: my-app-secretless-config
```

Great! Secretless Broker now has its configuration, and your app is up and running.

But what if you need to modify the configuration live, for example to add support for
a database connection? You can update the file in the ConfigMap and manually restart
the application. If instead you use the `configurations.secretless.io` Custom Resource
Definition (CRD), natural updates that occur in the course of evolving your application
are even easier to handle.

# The Kubernetes-Native Method - Configuration By Custom Resource

To use the configuration CRD, there is a one-time configuration that needs to
happen beforehand to make sure the Secretless CRD is installed in your cluster
and its assets are visible to the Secretless Broker. For more information on installing
the CRD, please see [our CRD documentation](https://github.com/cyberark/secretless-broker/blob/master/resource-definitions/README.md).

The deployment manifest for a CRD-based configuration looks similar to the `secretless.yml`
file we wrote above:

```
apiVersion: "secretless.io/v1"
kind: Configuration
metadata:
  name: my-app-secretless-config
spec:
  listeners:
    - name: my_webapp_listener
      protocol: http
      address: 0.0.0.0:8080

  handlers:
    - name: my_webapp_handler
      type: basic_auth
      listener: my_webapp_listener
      match:
        - ^http.*
      credentials:
        - name: username
          provider: environment
          id: WEBAPP_USERNAME
        - name: password
          provider: environment
          id: WEBAPP_PASSWORD                      
```

The difference here is just the addition of standard Kubernetes manifest fields to define
the resource object. We've called this specific configuration CRD object
`my-app-secretless-config`, and if we save the manifest above as `secretless-config.yml`
we can upload it to our cluster using the usual `kubectl apply` call:

```
kubectl apply -f ./secretless-config.yaml
```

Now the Secretless configuration for our app is available in the cluster, and we
can refer to it in our application manifest when we deploy our app:

```
containers:
  - name: my-app
    # deployment details for my-app
  - name: secretless-broker
    image: cyberark/secretless-broker:latest
    imagePullPolicy: IfNotPresent
    args: ["-config-mgr", "k8s/crd#my-app-secretless-config"]
```

Notice the difference between this snippet of the deployment manifest versus the
snippet above - we no longer have to mount the configuration into the Secretless
container, we just need to pass the Secretless Broker the `config-mgr` argument
that points to the specific CRD configuration we've uploaded.

And that's it! Once we deploy our app with its Secretless Broker sidecar, it will
have access to the configuration we specified in `secretless-config.yaml`. If we
add new Service Connectors to the Secretless configuration (for example, if
our app needs to also connect to a database), we just update the file and apply
the changes:

```
kubectl apply -f ./secretless-config.yaml
```

Secretless will automatically restart and update its configuration with the new
data.

For more information on our configuration CRD, you can check out its
[GitHub README](https://github.com/cyberark/secretless-broker/blob/master/resource-definitions/README.md)
or our [reference documentation](/docs/reference/config-managers/k8s/crd.html).
Stay tuned for more details on using the configuration CRD with our sidecar
injector for an even more seamless Kubernetes experience!
