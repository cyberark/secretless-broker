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
        image: gcr.io/google_containers/echoserver:1.1
        imagePullPolicy: Always
        env:
        - name: http_proxy
          value: "http://0.0.0.0:3000"
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
            value: "${CONJUR_VERSION}"
          - name: CONJUR_APPLIANCE_URL
            value: "${CONJUR_APPLIANCE_URL}"
          - name: CONJUR_AUTHN_URL
            value: "${CONJUR_AUTHN_URL}"
          - name: CONJUR_ACCOUNT
            value: "${CONJUR_ACCOUNT}"
          - name: CONJUR_AUTHN_LOGIN
            value: "${CONJUR_AUTHN_LOGIN}"
          - name: CONJUR_SSL_CERTIFICATE
            valueFrom:
              configMapKeyRef:
                name: conjur-cert # configurable ?
                key: ssl-certificate
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
