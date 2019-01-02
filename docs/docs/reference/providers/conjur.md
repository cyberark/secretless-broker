---
title: Credential Providers
id: conjur
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/providers/conjur.html
---

## CyberArk Conjur
The Conjur provider (`conjur`) populates credentials from an external
[Conjur](https://www.conjur.org) service.

### Secretless Broker Configuration
To use the Conjur provider, Secretless must be configured to authenticate with
Conjur. The Secretless Broker currently supports several methods of authenticating
with Conjur (activating the first non-empty method in this order):

- `CONJUR_AUTHN_LOGIN` and `CONJUR_AUTHN_API_KEY` environment variables
- `CONJUR_AUTHN_TOKEN_FILE` environment variable
- Conjur Kubernetes authenticator-based authentication
  
  In this mode Secretless behaves as an [authn-k8s-client](https://github.com/cyberark/conjur-authn-k8s-client) 
  and retrieves machine identity through orchestrator-facilitated attestation.
  + Requires `CONJUR_AUTHN_URL` environment variable contains `authn-k8s`
  + Requires identical configuration environment variables as [authn-k8s-client](https://github.com/cyberark/conjur-authn-k8s-client)
  + See [Conjur docs](https://docs.conjur.org/Latest/en/Content/Integrations/Kubernetes_deployApplicationApplication.htm) for additional information on configuration

Both methods also require `CONJUR_APPLIANCE_URL` and `CONJUR_ACCOUNT` to
be set in the environment of the Secretless Broker. You may optionally
also include any other configuration environment variables that are
allowed by the [Conjur Go Client Library](https://github.com/cyberark/conjur-api-go).

### Examples
``` yaml
listeners:
  - name: pg_listener
    protocol: pg
    address: 0.0.0.0:5432

handlers:
  - name: pg_handler
    listener: pg_listener
    credentials:
      - name: address
        provider: conjur
        id: postgres/my-service/address
      - name: username
        provider: conjur
        id: postgres/my-service/username
      - name: password
        provider: conjur
        id: postgres/my-service/password
```
