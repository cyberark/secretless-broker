# Secretless Broker Custom Resource Definitions (CRDs)

This folder contains most of the resources for dealing with CRDs in the context of
this codebase. CRDs are planned to be used for dynamically updating the broker
configuration across the cluster when the API is changed.

For more information about CRDs, you can find more information
[here](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)

# Pre-requisites

In order to work with this module, as well as if you are using the `crd_watcher`
and `crd_injector`, you will need to be sure you have installed the
[development prerequisites](../CONTRIBUTING.md#prerequisites).

# Examples

To manually create the CRD, let's first ensure that we have none already defined:
```
$ kubectl get crd
No resources found.
```

Now lets apply our CRD file template:
```
$ kubectl apply -f ./secretless-resource-definition.yaml
customresourcedefinition.apiextensions.k8s.io "configurations.secretless.io" created
```

Let's take a look at what we have:
```
$ kubectl get crd
NAME                           AGE
configurations.secretless.io   1m

$ kubectl get crd -o yaml
apiVersion: v1
items:
- apiVersion: apiextensions.k8s.io/v1beta1
  kind: CustomResourceDefinition
  metadata:
    annotations:
      ...
    creationTimestamp: 2018-08-17T15:39:40Z
    generation: 1
    name: configurations.secretless.io
...

```

Let us now add some configuration data:
```
$ kubectl apply -f ./sbconfig-example.yaml
configuration.secretless.io "secretless-example-config" created
```

We can now list and display our configuration:
```
$ kubectl get sbconfig
NAME                        AGE
secretless-example-config   1m

$ kubectl describe sbconfig secretless-example-config
Name:         secretless-example-config
Namespace:    default
...

```

The charm of CRDs is also that we can have multiple CRD resources so that
we can apply multiple configurations to our cluster:
```
$ kubectl apply -f ./sbconfig-example2-v1.yaml
configuration.secretless.io "secretless-example-config2" created

$ kubectl get sbconfig
NAME                         AGE
secretless-example-config    22s
secretless-example-config2   17s

$ kubectl describe sbconfig secretless-example-config2
Name:         secretless-example-config2
...
Spec:
  Handlers:
    Credentials:
      Id:        user1
      Name:      username
      Provider:  literal
      Id:        password1
      Name:      password
      Provider:  literal
    Listener:    http_config_1_listener
    Match:
      ^http.*
    Name:  http_config_1_handler
    Type:  basic_auth
  Listeners:
    Address:   0.0.0.0:8080
    Name:      http_config_1_listener
    Protocol:  http
```

Let's try updating only this config with a new one:
```
$ kubectl apply -f ./sbconfig-example2-v2.yaml
configuration.secretless.io "secretless-example-config2" configured

$ kubectl describe sbconfig secretless-example-config2
Name:         secretless-example-config2
...
Spec:
  Handlers:
    Credentials:
      Id:        user2
      Name:      username
      Provider:  literal
      Id:        password2
      Name:      password
      Provider:  literal
    Listener:    http_config_1_listener
    Match:
      ^http.*
    Name:  http_config_1_handler
    Type:  basic_auth
  Listeners:
    Address:   0.0.0.0:9090
    Name:      http_config_1_listener
    Protocol:  http
```

Now that we are done, we can clean up both our configurations and the CRD:
```
$ kubectl delete sbconfig secretless-example-config
configuration.secretless.io "secretless-example-config" deleted

$ kubectl delete sbconfig secretless-example-config2
configuration.secretless.io "secretless-example-config2" deleted

$ kubectl delete crd configurations.secretless.io
customresourcedefinition.apiextensions.k8s.io "configurations.secretless.io" deleted
```

# Adding CRD through code

This method is a bit more complicated, especially if it's run in-cluster due to needing to
have service account privileges but with that prerequisite, you can then use the `crd_injector.go`:

```
$ kubectl get crd
No resources found.

$ go run ./crd_injector.go
2018/08/17 10:55:35 Secretless CRD injector starting up...
2018/08/17 10:55:35 Using home dir config...
2018/08/17 10:55:35 Creating K8s client...
2018/08/17 10:55:35 Creating CRD...
2018/08/17 10:55:35 CRD was uccessfully added!
2018/08/17 10:55:35 Done!

$ kubectl get crd
NAME                           AGE
configurations.secretless.io   5s
```

# Watching CRD changes from code

This code requires you to have a service account privileges but with those,
you can use the `crd_watcher.go`:

In one terminal start the watcher after you have added a CRD definition:
```
$ go run ./crd_watcher.go
2018/08/21 16:05:46 Secretless CRD watcher starting up...
2018/08/21 16:05:46 Using home dir config...
2018/08/21 16:05:46 Available configs: 0
2018/08/21 16:05:46 Watching for changes...
```

Open another terminal and add a CRD definition:
```
$ kubectl apply -f ./sbconfig-example.yaml
configuration.secretless.io "secretless-example-config" created
```

Observe that the watcher has noticed the change:
```
2018/08/21 16:05:46 Watching for changes...
metadata:
  name: secretless-example-config
  generatename: ""
  namespace: default
  selflink: /apis/secretless.io/v1/namespaces/default/configurations/secretless-example-config
  ...
spec:
  listeners:
  ...
  handlers:
  ...
```

We can also add the second definion and swap the versions a few times:
```
$ kubectl apply -f sbconfig-example2-v1.yaml
configuration.secretless.io "secretless-example-config2" created

$ kubectl apply -f sbconfig-example2-v2.yaml
configuration.secretless.io "secretless-example-config2" configured

$ kubectl apply -f sbconfig-example2-v1.yaml
configuration.secretless.io "secretless-example-config2" configured
```

You should notice that we see changes in the output of the watcher:
```
2018/08/28 10:17:09 Add
2018/08/28 10:17:09 Add event:
metadata:
  name: secretless-example-config2
...
spec:
  listeners:
  - address: 0.0.0.0:8080
    name: http_config_1_listener
...

2018/08/28 10:17:15 Update
2018/08/28 10:17:15 Update event:
2018/08/28 10:17:15 Old:
metadata:
  name: secretless-example-config2
...
spec:
  listeners:
  - address: 0.0.0.0:9090
    name: http_config_1_listen
...

2018/08/28 10:17:13 Update
2018/08/28 10:17:13 Update event:
2018/08/28 10:17:13 Old:
metadata:
  name: secretless-example-config2
...
spec:
  listeners:
  - address: 0.0.0.0:8080
    name: http_config_1_listener
...
```

After you are done, don't forget to clean up your CRDs:
```
$ kubectl delete sbconfig secretless-example-config
configuration.secretless.io "secretless-example-config" deleted

$ kubectl delete sbconfig secretless-example-config2
configuration.secretless.io "secretless-example-config2" deleted

$ kubectl delete crd configurations.secretless.io
customresourcedefinition.apiextensions.k8s.io "configurations.secretless.io" deleted
```
