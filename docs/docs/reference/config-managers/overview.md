---
title: Configuration Managers
id: configmanagers
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/config-managers/overview.html
---

## Overview

Configuration Managers apply different strategies to provide Secretless Broker with its configurations
and ways to update them. The Secretless Broker comes built-in with a couple of different Configuration
Managers and a plugin API so that almost any configuration source can be easily integrated. Selection
of the specific configuration manager plugin is done through the `-config-mgr <manager_id>[#<config_spec>]`
CLI flag. If no specific configuration managers are specified, `configfile` plugin is used.

Note: Secretless Broker CLI is tightly integrated with `configfile` plugin so its use is different
from other plugins.

Currently built-in configuration manager plugins are:
- [Configuration File `configfile` manager](/docs/reference/config-managers/configfile.html)
- [Kubernetes CRD `k8s/crd` manager](/docs/reference/config-managers/k8s/crd.html)

## Examples

Generic CLI invocation
```
$ secretless-broker -config-mgr managername#configspec
```

Start broker with `secretless.yaml` in your current directory:
```
$ secretless-broker
```

Start broker with `custom-config.yaml` in `/foo` folder:
```
$ # CLI is the same since configfile plugin goes through a number of places to find the configuration
$ secretless-broker -f /foo/custom-config.yaml
```

Start broker with `custom-config.yaml` in folder `/foo` with `inotify` watch while also explicitly specifying the
`configfile` plugin:
```
$ secretless-broker -config-mgr configfile#/foo/custom-config.yaml?watch=true
```

Start broker with Kubernetes CRD config manager that will watch for changes in `test-secretless-crd` resource:
```
$ secretless-broker -config-mgr k8s/crd#test-secretless-crd
```
