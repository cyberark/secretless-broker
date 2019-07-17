#!/usr/bin/env bash

. ./env.sh

cat << EOL
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  labels:
    app: "${APP_NAME}"
  name: "${APP_NAME}"
  namespace: "${APP_NAMESPACE}"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: "${APP_NAME}"
  template:
    metadata:
      labels:
        app: "${APP_NAME}"
    spec:
      serviceAccountName: "${APP_SERVICE_ACCOUNT_NAME}"
      containers:
      - name: app
        image: mysql/mysql-server:5.7
        command: [ "sleep", "infinity" ]
        imagePullPolicy: Always
      - name: "${APP_AUTHENTICATION_CONTAINER_NAME}"
        image: cyberark/secretless-broker:latest
        imagePullPolicy: Always
        args: ["-f", "/etc/secretless/secretless.yml"]
        env:
          - name: MY_POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: MY_POD_NAMESPACE
            valueFrom:
              fieldRef:
                fieldPath: metadata.namespace
          - name: MY_POD_IP
            valueFrom:
              fieldRef:
                fieldPath: status.podIP
          - name: CONJUR_VERSION
            value: 5
          - name: CONJUR_APPLIANCE_URL
            value: "${CONJUR_APPLIANCE_URL}"
          - name: CONJUR_AUTHN_URL
            value: "${CONJUR_APPLIANCE_URL}/authn-k8s/${AUTHENTICATOR_ID}"
          - name: CONJUR_ACCOUNT
            value: "${CONJUR_ACCOUNT}"
          - name: CONJUR_AUTHN_LOGIN
            value: "host/conjur/authn-k8s/${AUTHENTICATOR_ID}/apps/${APP_NAMESPACE}/service_account/${APP_SERVICE_ACCOUNT_NAME}"
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              configMapKeyRef:
                name: conjur-cert # configurable ?
                key: ssl-certificate
        readinessProbe:
          httpGet:
            path: /ready
            port: 5335
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 2
          failureThreshold: 60
        volumeMounts:
          - mountPath: /etc/secretless
            name: config
            readOnly: true
      volumes:
        - name: config
          configMap:
            name: secretless-config
            defaultMode: 420
EOL
