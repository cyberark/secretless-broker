# Secretless Broker Custom Resource Definitions (CRDs)

This folder contains most of the resources for dealing with CRDs in the context of
this codebase. CRDs are planned to be used for dynamically updating the broker
configuration across the cluster when the API is changed.

# Examples

To manually create the CRD, let's fist ensure that we have none already defined:
```
$ kubectl get crd
No resources found.
```

Now lets apply our CRD file template:
```
$ kubectl apply -f ./secretless.yaml
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
$ kubectl apply -f ./sconfig-example.yaml
configuration.secretless.io "secretless-example-config" created
```

We can now list and display our configuration:
```
$ kubectl get sconfig
NAME                        AGE
secretless-example-config   1m

$ kubectl describe sconfig secretless-example-config
Name:         secretless-example-config
Namespace:    default
...

```

Now that we are done, we can clean up both our configuration and the CRD:
```
$ kubectl delete sconfig secretless-example-config
configuration.secretless.io "secretless-example-config" deleted

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
$ go run resource-definitions/crd_watcher.go
2018/08/21 16:05:46 Secretless CRD watcher starting up...
2018/08/21 16:05:46 Using home dir config...
2018/08/21 16:05:46 Available configs: 0
2018/08/21 16:05:46 Watching for changes...
```

Open another terminal and add a CRD definition:
```
$ kubectl apply -f resource-definitions/sconfig-example.yaml
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
