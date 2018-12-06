# Contributing to the Secretless Broker

Thanks for your interest in the Secretless Broker. Before contributing, please
take a moment to read and sign our <a href="https://github.com/cyberark/secretless-broker/blob/master/Contributing_OSS/CyberArk_Open_Source_Contributor_Agreement.pdf" download="secretless-broker_contributor_agreement">Contributor Agreement</a>.
This provides patent protection for all Secretless Broker users and allows CyberArk
to enforce its license terms. Please email a signed copy to
<a href="oss@cyberark.com">oss@cyberark.com</a>.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Pull Request Workflow](#pull-request-workflow)
- [Style Guide](#style-guide)
- [Building](#building)
- [Testing](#testing)
- [Plugins](#plugins)
- [Releasing](#releasing)

## Prerequisites

### Mercurial (`hg`)
Due to a dependency on `k8s/client-go`, our project requires that you have
installed Mercurial (`hg` on the CLI) on your system.

macOS:
```
$ brew install mercurial
```

Linux:
```
# Alpine
$ apk add -u mercurial

# Debian-based
$ apt update
$ apt install mercurial
```

### A note on go.sum fixes for `k8s/client-go`

`k8s/client-go` downloads a package with mismatching `go.sum` which may present itself
as something like this during builds or attempts to run the code:
```
go: verifying k8s.io/client-go@v0.0.0-20180806134042-1f13a808da65: checksum mismatch
    downloaded: h1:wQUEIVcXYxsDE8RXfUufo1nfnkeH/BEPhT175YIzea4=
    go.sum:     h1:3w7osyUaXe5a1wxJrqkfjRhqYMfi9pCiB64J9bmtszk=
```

If you see this problem, you need to remove the `k8s/client-go` checksum from the
repository-provided file with the following code and retry your build/run command:

```
sed -i '/^k8s.io\/client-go\ /d' go.sum
```

In general, we get around this problem for now by editing the go.sum lines related
to `k8s.io/client-go` in the Secretless Broker Dockerfiles.

## Pull Request Workflow

1. Search the [open issues][issues] in GitHub to find out what has been planned
2. Select an existing issue or open an issue to propose changes or fixes
3. Move the issue to "in progress" in [Waffle][waffle] as you work on it
4. Run tests as described [here][tests], ensuring they pass
5. Submit a pull request, linking the issue in the description
6. Move the issue to "in review" in [Waffle][waffle], ask another contributor to review and merge your code

Our [Waffle.io][waffle] is synchronized with GitHub and helps you navigate this workflow more easily.

In addition to technical workflow descriptions available in Waffle / GitHub,
some of the project's technical design documents can be found in the project [design][design] folder.

## Style guide

Use [this guide][style] to maintain consistent style across the Secretless Broker project.

[design]: https://github.com/cyberark/secretless-broker/tree/master/design
[issues]: https://github.com/cyberark/secretless-broker/issues
[style]: STYLE.md
[tests]: #testing
[waffle]: https://waffle.io/cyberark/secretless

## Building

First, clone `https://github.com/cyberark/secretless-broker`. If you're new to Go, be aware that Go can be very selective
about where the files are placed on the filesystem. There is an environment variable called `GOPATH`, whose default value
is `~/go`. Secretless Broker uses [go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) which
require either that you clone this repository outside of your `GOPATH` or you set the `GO111MODULE` environment variable to
`on`. We recommend cloning this repository outside of your `GOPATH`.

Once you've cloned the repository, you can build the Secretless Broker.

### Docker containers

```sh-session
$ # From Secretless Broker repository root
$ ./bin/build
```

This should create a Docker container with tag `secretless-broker:latest` in your local registry.

### Binaries
#### Linux
```sh-session
$ # From Secretless Broker repository root
$ go build -o ./secretless-broker ./cmd/secretless-broker
```

#### OSX

```sh-session
$ # From Secretless Broker repository root
$ ./bin/build_darwin
```

## Testing

**Prerequisites**

* **Docker** You need Docker to run the tests.

Build the project by running:

```sh-session
$ ./bin/build
```

Then run the test cases:

```sh-session
$ ./bin/test
```

If you are on a Mac, you may also test the OSX Keychain provider:
```sh-session
cd test/manual/keychain_provider/
./start
./test
```
This test will not be run as part of the test suite, since it requires access
to the Mac OSX Keychain. You will be prompted for your password when running
this test, as it temporarily adds a generic password to your account, and
verifies that it can retrieve the value.

Kubernetes CRD loading test
```sh-session
cd test/manual/k8s_crds
./deploy
```
This test currently does not run as part of the test suite.

## Plugins

Plugins can be used to extend the functionality of the Secretless Broker via a shared library in `/usr/local/lib/secretless` by providing a way to add additional:

- Listener plugins
- Handler plugins
- Connection management plugins

Currently, these API definitions reside [here](pkg/secretless/plugin/v1) and an example plugin can be found in the [`test/plugin`](test/plugin) directory.

You can read more about how to make plugins and the underlying architecture in the [API directory](pkg/secretless/plugin).

_Please note: Plugin API interface signatures and supported plugin API version(s) are currently under heavy development so they will be likely to change in the near future._

## Releasing

1. Based on the unreleased content, determine the new version number and update
   the [VERSION](VERSION) file.
1. Run `./bin/prefill_changelog` to populate the [changelog](CHANGELOG.md) with
   the changes included in the release.
1. Commit these changes - `Bump version to x.y.z` is an acceptable commit message.
1. Once your changes have been reviewed and merged into master, tag the version
   using `git tag -s v0.1.1`. Note this requires you to be  able to sign releases.
   Consult the [github documentation on signing commits](https://help.github.com/articles/signing-commits-with-gpg/)
   on how to set this up. `vx.y.z` is an acceptable tag message.
1. Push the tag: `git push vx.y.z` (or `git push origin vx.y.z` if you are working
   from your local machine).
1. From a **clean checkout of master** run `./bin/build_release` to generate
   the release artifacts. Upload these to the GitHub release.
