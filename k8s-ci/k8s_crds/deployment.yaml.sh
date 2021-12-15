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

---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: configurations.secretless${SECRETLESS_CRD_SUFFIX}.io
spec:
  group: secretless${SECRETLESS_CRD_SUFFIX}.io
  names:
    kind: Configuration
    plural: configurations
    singular: configuration
    shortNames:
      - sbconfig
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                listeners:
                  type: array
                  items:
                    type: object
                    properties:
                      name:
                        type: string
                      protocol:
                        type: string
                      socket:
                        type: string
                      address:
                        type: string
                      debug:
                        type: boolean
                      caCertFiles:
                        type: array
                        items:
                          type: string
                handlers:
                  type: array
                  items:
                    type: object
                    properties:
                      name:
                        type: string
                      type:
                        type: string
                      listener:
                        type: string
                      debug:
                        type: boolean
                      match:
                        type: array
                        items:
                          type: string
                      credentials:
                        type: array
                        items:
                          type: object
                          properties:
                            name:
                              type: string
                            provider:
                              type: string
                            id:
                              type: string

EOL
