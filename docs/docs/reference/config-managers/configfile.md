---
title: Configuration Managers
id: configfile
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/config-managers/configfile.html
---

## Configuration File
The configuration file plugin (`configfile`) allows the use of parsing a YAML file as
the source of Secretless Broker configuration. Most of the functionality of this plugin
is abstracted away through common CLI arguments like `-f <path>` and `-watch` so no
special handling of this type of configuration is needed.

By default, if no `-f` parameters are specified, the configuration plugin goes through the
following steps:

- Load the configuration file
  - Try to load `./secretless.yml`
  - Try to load `$HOME/.secretless.yml` if the previous step fails
  - Try to load `/etc/secretless.yml` if the previous step fails
  - If none of the files we tried to previously load worked, exit with a failure
- Apply `inotify` filesystem watch for changes if the `-watch` flag is specified

If `-f <file>` parameter is specified, the following steps are taken:

- Load the configuration from `<file>`
- Fail startup if the file cannot be found
- Apply `inotify` filesystem watch for changes if the `-watch` flag is specified

## Examples

Start broker with `secretless.yaml` in your current directory:
```
$ secretless-broker
```

Start broker with `custom-config.yaml` in `/foo` folder:
```
$ secretless-broker -f /foo/custom-config.yaml
```

Start broker with `custom-config.yaml` in `/foo` folder and reload the configuration if the file
content changes:
```
$ secretless-broker -watch -f /foo/custom-config.yaml
```

Start broker with `custom-config.yaml` in folder `/foo` with `inotify` watch while also explicitly specifying the
`configfile` plugin:
```
$ secretless-broker -config-mgr configfile#/foo/custom-config.yaml?watch=true
```
