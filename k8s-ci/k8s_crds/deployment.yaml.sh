#!/bin/bash

set -euo pipefail

# returns current namespace if available, otherwise returns 'default'
current_namespace() {
  cur_ctx="$(kubectl config current-context)" || exit_err "error getting current context"
  ns="$(kubectl config view -o=jsonpath="{.contexts[?(@.name==\"${cur_ctx}\")].context.namespace}")" \
     || exit_err "error getting current namespace"

  if [[ -z "${ns}" ]]; then
    echo "default"
  else
    echo "${ns}"
  fi
}

cat << EOL
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
  - "secretless${SECRETLESS_CRD_SUFFIX}.io"
  resources:
  - configurations
  verbs:
  - get
  - list
  - watch

---
apiVersion: v1
kind: ServiceAccount
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
  namespace: $(current_namespace)
roleRef:
  kind: ClusterRole
  name: secretless-crd
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: secretless-k8s-crd-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: secretless-k8s-crd-test
  template:
    metadata:
      labels:
        app: secretless-k8s-crd-test
    spec:
      serviceAccountName: secretless-crd
      containers:
      - name: echo-server
        image: gcr.io/google_containers/echoserver:1.10
        imagePullPolicy: Always
      - name: secretless
        args: [ "-config-mgr", "k8s/crd#first" ]
        env:
        - name: SECRETLESS_CRD_SUFFIX
          value: "${SECRETLESS_CRD_SUFFIX}"
        image: "${SECRETLESS_IMAGE}"
        readinessProbe:
          tcpSocket:
            port: 8080
          initialDelaySeconds: 15
          timeoutSeconds: 5
        imagePullPolicy: Always
        ports:
          - containerPort: 8080
EOL
