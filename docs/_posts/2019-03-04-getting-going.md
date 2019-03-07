---
layout: post
title: "Getting GOing"
date: 2019-03-04 09:00:00 -0600
author: Sigal Sax
categories: blog
published: true
image: go-gopher-illustration.jpg
thumb: go-gopher-illustration.jpg
image-alt: Go Gopher illustration
excerpt: "Why we chose Go for the Secretless Broker, and how this decision has impacted our development process."
---

From the very beginning, the original Golang developers had a clear goal - retain the positive attributes of the common programming languages while building one that was readable, simple, and high performing. Go’s light-weight model for fast compilation, extensive toolset, and native capabilities are what have propelled it toward worldwide adoption and success.  With these modern features and being backed by Google, it is no wonder why open source projects such as Docker, Kubernetes, and some in CyberArk, are being built in Go.  

## Get GOing with Secretless  

For the past year, we on the CyberArk Conjur team have been working on our latest open source project in Go, the Secretless Broker. The Secretless Broker acts as a transparent proxy that injects the necessary credentials to target services directly, relieving client applications of the responsibility of handling sensitive data. Essentially, you can open a local connection to a credentialed resource without ever having to pass credentials to do so; all with Secretless. By removing the direct interaction between applications and target resources, the potential for a secret leak decreases significantly. 

I was new to this project and especially new to the Go community, so naturally, I was curious why we made the design decision to build the Secretless Broker in Go. After working on features, participating in design discussions and debates during team stand-ups, and having just attended the inaugural [GopherCon Israel](https://www.gophercon.org.il/), it became clear why. We needed a system-level language that was as fast as the C-family but not as low level. Go gave us the perks of a low-level language while providing us with a toolset and native capabilities of a higher-level one.

Go offers a full suite of native capabilities rarely packaged together in other programming languages. Go’s static typing, high-level APIs, and cross-compilation support are just a few features that have aided our development process. We found that Go, being statically typed, has saved us countless hours of tracing and debugging. If at compile-time our statically-typed program runs, we can be more confident our program won’t break during execution.  

Go's high-level APIs have also been valuable to running faster development cycles. Usually in a low-level language, it can be a real pain to write hundreds of lines of code just to implement a certain function. With the high-level APIs that Go provides, most of this is concealed through abstraction.  

## The toolchain 

Go’s open source tools and built-in toolchain add additional functionality to projects. These tools can either be imported or implemented into any code environment as stand-alone console applications. Specifically, for the Secretless Broker, the tools we enjoy and find most helpful are `pprof` for profiling, `golint` for code clean up, and `godocs` for documentation automation. 

`pprof`is a profiling tool that once added to the main function of a program, provides an in-depth analysis of processes. With Secretless, every byte of data travels through the broker before reaching and returning to its endpoints. We found that this exchange was taking too long, but we couldn’t tell exactly where the latency was coming from. With pprof, we were able visualize the effect of our project on a host machine’s CPU and memory. This is crucial because by using this tool, we gain visibility into processes and can tune them to provide a more positive experience for our customers. 

It also became difficult to consistently format our code, especially as our codebase increased in size and complexity. To fix this common problem, we added the `golint` automated tool into our workflow. Code linters run source code against an applied set of stylistic and syntax rules, raising flags when this structure is broken.  Linting has been an indispensable productivity tool for writing code and has even sped up our code review process.  

Documentation is essential to growing and sustaining the community around our project. We add to and tweak our README and contributor guides to encourage end-user self-efficiency and success. This has proven to be a time-consuming undertaking because our codebase is constantly evolving and being refactored. To our benefit, Go offers the `godoc` documentation tool to dynamically create documentation. This tool parses the code, relevant comments, and structure to generate documentation that evolves as fast as our code does. 

## … and beyond the toolchain 

In the early stages of our project, we have been focusing on providing a native and secure Secretless solution for Kubernetes and Openshift environments. To accomplish this, we wanted to deliver Secretless so that it could be deployed as a sidecar container. Secretless is consumed as a Linux binary that is packaged inside a Docker image. That way, customers can easily get up and running once they grab our image. We also use Docker for our CI pipeline to build and publish our site, check styling and links, and run tests. We host these processes in Docker to ensure they receive the proper Go versioning for testing and to isolate their environments for reproducible results. 

When it comes to testing, I have mixed feelings about Docker. Docker has added an extra layer of complexity and an additional prerequisite for contributing that may make our project less approachable. In terms of testing norms, Docker-based development has forced us to break the standard Go testing conventions. Our tests have become custom to our project, wrapping the known Go tests with Docker-specific code. Consequently, it may become initially unclear to contributors how our tests work.  However, in using Docker, we are able to run integration tests that mimic a more realistic Secretless deployment, with Secretless running in a container and communicating with a database service, for example. Since Docker provides pre-built images and ensures complete process isolation, it makes the setup of test environments a breeze. We find Docker beneficial because the alternative is a large assortment of messy Bash scripts and boilerplate code. Furthermore, contributors no longer need to fuss with tooling in order to configure their environment correctly - all they need is Docker and they are ready to get rolling.

To balance these complexities and ensure a positive contributor experience, we are working to enable standard local development while still leveraging the benefits of Docker when appropriate, and trying to provide better documentation so it's easy to get started. We hope this will eventually give our contributors the support and guidance they need to get involved and continue to contribute to our project.

## Summing Up 

Go is trending in the open source world and its support and community are only growing. For our project, we found Go’s native capabilities and supportive toolset to be specifically helpful as our project grew in complexity. Go’s sophistication and modern approach to computing are what have ultimately allowed us to stretch outside the conventions of a normal open source project.  

Looking to [contribute to Secretless?](https://github.com/cyberark/secretless-broker/blob/master/CONTRIBUTING.md) We would love nothing more!  
