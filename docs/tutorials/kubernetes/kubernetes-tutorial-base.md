---
title: Using Secretless in Kubernetes
id: kubernetes_tutorial
layout: tutorials
description: Secretless Broker Documentation
section-header: Before we begin...
time-complete: 5
products-used: Kubernetes Secrets, PostgreSQL Service Connector
back-btn:
continue-btn: /tutorials/kubernetes/overview.html
up-next: Get an overview of what is going to be covered in this tutorial.
permalink: /tutorials/kubernetes/kubernetes-tutorial-base.html
---
This is a detailed, step-by-step tutorial.

You will:

1. Deploy a PostgreSQL database
2. Store its credentials in Kubernetes secrets
3. Setup Secretless Broker to proxy connections to it
4. Deploy an application that connects to the database **without knowing its password**

Already a Kubernetes expert? You may prefer our <a href="https://github.com/cyberark/secretless-broker/tree/master/demos/k8s-demo">advanced Github tutorial</a> complete with shell scripts to get **the whole thing working end to end fast**.
