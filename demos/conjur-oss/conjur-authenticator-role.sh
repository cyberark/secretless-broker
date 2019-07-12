#!/usr/bin/env bash

. ./env.sh

cat << EOL
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: conjur-authenticator
  namespace: ${APP_NAMESPACE}
rules:
- apiGroups: [""] # "" indicates the core API group
  resources: ["pods", "serviceaccounts"]
  verbs: ["get", "list"]
- apiGroups: ["extensions"]
  resources: [ "deployments", "replicasets"]
  verbs: ["get", "list"]
- apiGroups: ["apps"]  # needed on OpenShift 3.7+
  resources: [ "deployments", "statefulsets", "replicasets"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create", "get"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: conjur-authenticator-role-binding
  namespace: ${APP_NAMESPACE}
subjects:
  - kind: ServiceAccount
    name: ${OSS_CONJUR_SERVICE_ACCOUNT_NAME}
    namespace: ${OSS_CONJUR_NAMESPACE}
roleRef:
  kind: Role
  name: conjur-authenticator
  apiGroup: rbac.authorization.k8s.io
EOL
