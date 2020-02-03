---
title: Using Secretless in Kubernetes
id: kubernetes_tutorial
layout: tutorials
description: Secretless Broker Documentation
section-header: Steps for Application Developer
time-complete: 5
products-used: Kubernetes Secrets, PostgreSQL Service Connector
back-btn: /tutorials/kubernetes/sec-admin.html
continue-btn: /tutorials/kubernetes/appendix.html
up-next: Get a closer look at Secretless...
permalink: /tutorials/kubernetes/app-dev.html
---

<div class="change-role">
  <div class="character-icon"><img src="/img/application_developer.jpg" alt="Application Developer"/></div>
  <div class="content">
    <div class="change-announcement">
      You are now the application developer.  
    </div>
    <div class="message">
      You can no longer access the secrets we stored previously in environment
      variables.  Open a new terminal so that all those variables are gone.
    </div>
  </div>
</div>

You know only one thing -- the name of the database:

```bash
export APPLICATION_DB_NAME=quick_start_db
```

### Sample Application Overview

The application we'll be deploying is a [pet store demo
application](https://github.com/conjurdemos/pet-store-demo) with a simple API:

- `GET /pets` lists all the pets
- `POST /pet` adds a pet

Its PostgreSQL backend is configured using a `DB_URL` environment variable:

<pre>
postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
</pre>

Again, the application has no knowledge of the database credentials it's using.

For usage examples, please see [Test the Application](#test-the-application).

### Create Application Deployment Manifest

We're ready to deploy our application.

A detailed explanation of the manifest below is featured in the next step, <a href="/tutorials/kubernetes/appendix.html">Appendix - Secretless
Deployment Manifest Explained</a> and isn't needed to complete the tutorial.

To create the **quick-start-application.yml** manifest using the
`APPLICATION_DB_NAME` above, run:

```bash
cat << EOF > quick-start-application.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quick-start-application
  labels:
    app: quick-start-application
spec:
  replicas: 3
  selector:
    matchLabels:
      app: quick-start-application
  template:
    metadata:
      labels:
        app: quick-start-application
    spec:
      serviceAccountName: quick-start-application
      automountServiceAccountToken: true
      containers:
        - name: quick-start-application
          image: cyberark/demo-app:latest
          env:
            - name: DB_URL
              value: postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
        - name: secretless-broker
          image: cyberark/secretless-broker:latest
          imagePullPolicy: Always
          args: ["-f", "/etc/secretless/secretless.yml"]
          volumeMounts:
            - name: config
              mountPath: /etc/secretless
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: quick-start-application-secretless-config
EOF
```

### Deploy Application With Secretless Broker

To deploy the application, run:

```bash
kubectl --namespace quick-start-application-ns apply -f quick-start-application.yml
```
<pre>
deployment "quick-start-application" created
</pre>

Before moving on, verify that the pods are healthy:

```bash
kubectl --namespace quick-start-application-ns get pods
```
<pre>
NAME                                       READY     STATUS        RESTARTS   AGE
quick-start-application-6bd8dbd57f-bshmf   2/2       Running       0          22s
quick-start-application-6bd8dbd57f-dr962   2/2       Running       0          26s
quick-start-application-6bd8dbd57f-fgfnh   2/2       Running       0          30s
</pre>

### Expose Application Publicly

The application is running, but not yet publicly available.

To expose it publicly as a Kubernetes Service, run:

```bash
cat << EOF > quick-start-application-service.yml
kind: Service
apiVersion: v1
metadata:
  name: quick-start-application
spec:
  selector:
    app: quick-start-application
  ports:
  - port: 8080
    targetPort: 8080
    nodePort: 30002
  type: NodePort
EOF
kubectl --namespace quick-start-application-ns \
 apply -f quick-start-application-service.yml
```
<pre>
service "quick-start-application" created
</pre>

Congratulations!

The application is now available at `$(minikube ip):30002`.  We'll call
this the `APPLICATION_URL` going forward.

## Test the Application

Let's verify everything works as expected.

First, make sure the `APPLICATION_URL` is correctly set:

```bash
export APPLICATION_URL=$(minikube ip):30002
```

Now let's create a pet (`POST /pet`):

```bash
curl -i -d @- \
 -H "Content-Type: application/json" \
 ${APPLICATION_URL}/pet \
 << EOF
{
   "name": "Secretlessly Fluffy"
}
EOF
```
<pre>
HTTP/1.1 201
Location: http://192.168.99.100:30002/pet/2
Content-Length: 0
Date: Thu, 23 Aug 2018 11:56:27 GMT
</pre>

We should get a 201 response status.

Now let's retrieve all the pets (`GET /pets`):

```bash
curl -i ${APPLICATION_URL}/pets
```
<pre>
HTTP/1.1 200
Content-Type: application/json;charset=UTF-8
Transfer-Encoding: chunked
Date: Thu, 23 Aug 2018 11:57:17 GMT

[{"id":1,"name":"Secretlessly Fluffy"}]
</pre>

We should get a 200 response with a JSON array of the pets.

That's it!

<div class="the-big-finish">
  <p>
  The application is connecting to a password-protected Postgres database
  <b>without any knowledge of the credentials</b>.
  </p>

  <img src="/img/its_magic.jpg" class="k8s-img" alt="It's Magic"/>
</div>

For more info on configuring Secretless for your own use case, see the <a href="https://docs.secretless.io/Latest/en/Content/Overview/scl_how_it_works.htm">Secretless Documentation</a>
