---
title: Using Secretless in Kubernetes
id: overview
layout: tutorial
description: Secretless Broker Documentation
permalink: /docs/tutorial_slides/intro.html
---

This is a detailed, step-by-step tutorial.

You will:

1. Deploy a PostgreSQL database
2. Store its credentials in Kubernetes secrets
3. Setup Secretless Broker to proxy connections to it
4. Deploy an application that connects to the database **without knowing its password**

Already a Kubernetes expert? You may prefer our:

<div style="text-align: center">
  <a href="https://github.com/cyberark/secretless-broker/tree/master/demos/k8s-demo" class="button btn-primary gradient">Advanced Github Tutorial</a>
</div>

complete with shell scripts to get **the whole thing working end to end fast**.

## Table of Contents

+ [Overview](#overview)
+ Steps for Security Admin
  + [Create PostgreSQL Service in Kubernetes](#create-postgresql-service-in-kubernetes)
  + [Create Application Database](#create-application-database)
  + [Create Application Namespace and Store Credentials](#create-application-namespace-and-store-credentials)
  + [Create Secretless Broker Configuration ConfigMap](#create-secretless-broker-configuration-configmap)
  + [Create Application Service Account and Grant Entitlements](#create-application-service-account-and-grant-entitlements)
+ Steps for Application Developer
  + [Sample Application Overview](#sample-application-overview)
  + [Create Application Deployment Manifest](#create-application-deployment-manifest)
  + [Deploy Application With Secretless Broker](#deploy-application-with-secretless-broker)
  + [Expose Application Publicly](#expose-application-publicly)
+ [Test the Application](#test-the-application)
+ [Appendix - Secretless Deployment Manifest Explained](#appendix---secretless-deployment-manifest-explained)
  + [Networking](#networking)
  + [SSL](#ssl)
  + [Credential Access](#credential-access)
  + [Configuration Access](#configuration-access)
