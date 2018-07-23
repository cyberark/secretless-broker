---
title: Secretless
id: quick_start
layout: docs
description: Secretless Documentation
permalink: quick_start
---

<p class="card-heading">Quick Start</p>

# Introduction

In this guide, we'll demonstrate Secretless brokering access to a PostgreSQL database using Docker. If you're looking for instructions on running Secretless within Kubernetes, visit our <a href="/deploy_to_kubernetes.html">Deploying to Kubernetes</a> demo.

# Running the demo environment

docker-compose.yml
```
version: "2"

services:
  postgres:
    image: postgres:9.6

  secretless:
    image: cyberark/secretless:latest
    ports:
      - 5432:15432
    volumes:
      - "./secretless.yml:/etc/secretless.yml:ro"
    command: [ "-f", "/etc/secretless.yml" ]
```

secretless.yml
```
listeners:
  - name: pg_tcp
    protocol: pg
    address: 0.0.0.0:15432

handlers:
  - name: pg
    listener: pg_tcp
    credentials:
      - name: address
        provider: literal
        id: postgres:5432
      - name: username
        provider: literal
        id: postgres
      - name: password
        provider: literal
        id: postgres
```
