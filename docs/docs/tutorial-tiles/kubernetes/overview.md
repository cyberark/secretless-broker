---
title: Using Secretless in Kubernetes
id: kubernetes_tutorial
layout: tutorials
description: Secretless Broker Documentation
section-header: Overview
time-complete: 5
products-used: Conjur Vault, Kubernetes
back-btn: /docs/tutorial-tiles/kubernetes/base.html
continue-btn: /docs/tutorial-tiles/kubernetes/sec-admin.html
up-next: Get an overview of what is going to be covered in this tutorial.
permalink: /docs/tutorial-tiles/kubernetes/overview.html
---
Applications and application developers should be **incapable of leaking secrets**.

To achieve that goal, you'll play two roles in this tutorial:

1. A **Security Admin** who handles secrets, and has sole access to those secrets
2. An **Application Developer** with no access to secrets.

The situation looks like this:

![Image](/img/secretless_overview.jpg)

Specifically, we will:

**As the security admin:**

1. Create a PostgreSQL database
1. Create a DB user for the application
1. Add that user's credentials to Kubernetes Secrets
1. Configure Secretless to connect to PostgreSQL using those credentials

**As the application developer:**

1. Configure the application to connect to PostgreSQL via Secretless
1. Deploy the application and the Secretless sidecar

### Prerequisites

To run through this tutorial, you will need:

+ A running Kubernetes cluster (you can use
  [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) to run a
  cluster locally)
+ [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) configured
  to point to the cluster
+ [Docker CLI](https://docs.docker.com/install/)
