---
layout: post
title: "Adding Support For A New Target Service"
date: 2018-09-20 09:00:00 -0600
author: Geri Jennings
categories: blog
published: true
image: secretless_logo_blog.jpg
thumb: secretless_logo_blog.jpg
image-alt: Secretless logo
excerpt: "Using Secretless Broker's Built-In Plugin Architecture to Add Features"
---

Secretless Broker has [documentation](/generated/pkg_secretless_plugin_v1.html)
on its plugin architecture, but to make it as easy as possible to contribute
new functionality we'll spend some time in this blog post breaking down how to
contribute support for a new Target Service.

<img src="/img/secretless_internal_architecture.svg" alt="Secretless Broker Internal Architecture">

In our [reference](https://docs.secretless.io/Latest/en/Content/Overview/scl_how_it_works.htm), we break down how the
Secretless Broker internal architecture handles incoming requests. Every target
service that Secretless Broker natively supports has its own Service Connector
implemented in the Secretless internals.

  - The Service Connector listens on a TCP port or Unix socket for incoming connections
  - The Service Connector uses standard functionality to retrieve the credentials it needs
    and opens a connection to the Target Service with those credentials injected
  - The Service Connector streams the connection

In what follows, we'll walk through the steps you would take to add a new
Service Connector to the Secretless project. We'll focus on adding it to the
project internals, but at the end we'll briefly give some guidance for how to build an
external plugin into the broker binary.

# Adding a New Listener

The first step in adding support to Secretless Broker for your service is to add
a `newservice` folder to the [/internal/app/secretless/listeners/](https://github.com/cyberark/secretless-broker/tree/master/internal/app/secretless/listeners)
directory. In that directory you'll create a file `listener.go` that has to
implement
