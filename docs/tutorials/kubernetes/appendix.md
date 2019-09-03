---
title: Using Secretless in Kubernetes
id: kubernetes_tutorial
layout: tutorials
description: Secretless Broker Documentation
section-header: Appendix - Secretless Deployment Manifest Explained
time-complete: 5
products-used: Kubernetes Secrets, PostgreSQL Service Connector
back-btn: /tutorials/kubernetes/app-dev.html
continue-btn: /tutorials/kubernetes/finish.html
up-next: A summary of what you accomplished in this tutorial!
permalink: /tutorials/kubernetes/appendix.html
---
Here we'll walk through the application deployment manifest, to better
understand how Secretless works.

We'll focus on the Pod's template, which is where the magic happens:

```yaml
  # top part elided...
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
```

### Networking

Since it resides in the same pod, the application can access the Secretless
sidecar container over localhost.

As specified in the ConfigMap we created, Secretless listens on port
`5432`, and hence this:

```yaml
          env:
            - name: DB_URL
              value: postgresql://localhost:5432/${APPLICATION_DB_NAME}?sslmode=disable
```

is all our application needs to locate Secretless.

### SSL

Notice the `?sslmode=disable` at the end of our `DB_URL`.

This means that **the application connects to Secretless without SSL**, which
is safe because it is intra-Pod communication over localhost.

However, the **connection between Secretless and Postgres is secure, and does
use SSL**.  

The situation looks like this:

```
                 No SSL                       SSL
Application   <---------->   Secretless   <---------->   Postgres
```

For more information on PostgreSQL SSL modes see:

- [PostgreSQL SSL documentation](https://www.postgresql.org/docs/9.6/libpq-ssl.html)
- [PostgreSQL Secretless Service Connector documentation](https://docs.secretless.io/Latest/en/Content/References/connectors/postgres.htm).

### Credential Access

Notice we add the **quick-start-application** ServiceAccount to the pod:

```yaml
    spec:
      serviceAccountName: quick-start-application
```

That's the ServiceAccount we created earlier, the one with access to the
credentials in Kubernetes Secrets.  This is what gives Secretless access
to those credentials.

### Configuration Access

Finally, notice the sections defining the volumes and the volume mount in the
Secretless container:

```yaml
          # ... elided
          volumeMounts:
            - name: config
              mountPath: /etc/secretless
              readOnly: true
      volumes:
        - name: config
          configMap:
            name: quick-start-application-secretless-config
```

Here we create a volume base on the ConfigMap we created earlier, which stores
our **secretless.yml** configuration file.

Thus Secretless gets its configuration file via a volume mount.
