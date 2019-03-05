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
- [Documentation](#documentation)
- [Profiling](#profiling)
- [Plugins](#plugins)
- [Releasing](#releasing)

## Prerequisites

### Go version
To work in this codebase, you will want to have at least Go 1.11.4 installed.

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

## Documentation
Secretless has a few sources for documentation: a website, a documentation subdomain, and godocs.

### Website
The [website](https://secretless.io) source is in the [docs](docs/) folder in this repository. It is generated using Jekyll.

#### Prerequisites

To get the site up and running locally on your computer, ensure you have:
1. Ruby version 2.1.0 or higher (check by running `ruby -v`)
2. Bundler (`gem install bundler`)
3. Jekyll (`gem install jekyll`)
4. Once Bundler and Jekyll gems are installed, run `bundle install`

#### Run Locally
To construct:
1. `git clone https://github.com/cyberark/secretless-broker`
2. `cd docs`
3. Run the following command:
`bundle exec jekyll serve`
4. Preview Jekyll site locally in web browser by either running `open localhost:4000` or manually navigating to http://localhost:4000

#### Run in Docker
With `docker` and `docker-compose`:

1. Run `docker-compose up -d` in the `docs` directory.
2. Preview Jekyll site locally in web browser by either running `open localhost:4000` or manually navigating to http://localhost:4000

### Documentation Website
The [documentation website](https://docs.secretless.io) source is in the [secretless-docs repo](https://github.com/cyberark/secretless-docs); instructions for contributing are available there.

### Godocs
[Godocs](https://godoc.org/github.com/cyberark/secretless-broker) are auto-published. To preview godocs locally, run `./bin/dev_godocs` and visit `localhost:6060` in your browser - changes you make locally will be available in your browser as you refresh the page.

## Profiling
Profiling can be used to monitor the impact of Secretless on CPU and Memory consumption. Currently, Secretless supports two types- CPU and Memory.

**Prerequisites:**
- [Graphviz](https://graphviz.gitlab.io/download/) to visualize profiling results
- [Postgresql](https://www.postgresql.org/download/) to install Postgres

We've provided sample instructions below for profiling the PostgreSQL handler.

*Note: If you are running through these instructions yourself, you'll want to
replace `<GOOS>/<GOARCH>` with your particular operating system and compilation architecture.*

1. [Build](#building) Secretless locally
1. Run a Postgres backend named ```sample-pg```:
   ```
   pushd test/pg_handler
     docker build -t sample-pg -f Dockerfile.pg .
     docker run -d -p 5432:5432 sample-pg
   popd
   ```

1. Check if Postgres is running and query the database:
   ```
   $ psql -h localhost -p 5432 -U test dbname=postgres -c "select count(*) from test.test;"
    count  
   --------
    100000
   (1 row)
   ```

1. Create a sample secretless.yml file in the project root that has:

   ```
   listeners:
     - name: pg_tcp
       protocol: pg
       address: 0.0.0.0:15432

   handlers:
     - name: pg_via_tcp
       listener: pg_tcp
       credentials:
         - name: address
           provider: env
           id: PG_ADDRESS
         - name: username
           provider: literal
           id: test
         - name: password
           provider: env
           id: PG_PASSWORD
    ```

1. The type of profiling is explicitly defined in the initial command that runs Secretless. Run Secretless with the profile desired like so:
   ```
   $ PG_ADDRESS=localhost:5432/postgres \
       PG_PASSWORD=test \
       ./dist/<GOOS>/<GOARCH>/secretless-broker \
       -profile=<cpu or memory> \
       -f secretless.yml
   ```
   *Note: The location of the binary may vary across different OS*

1. Once Secretless is running, the type of profiling defined in the previous step should state that it has been enabled. It should look something like:
   ```
   2018/11/21 10:17:13 profile: cpu profiling enabled, /var/folders/wy/f9qn852d5_d4s_g06s1kwjcr0000gn/T/profile789228879/cpu.pprof
   ```
   *Note: The hash observed will be different each time a profile is run.*

1. Once the Postgres database and Secretless are spun up, query the database through Secretless by running the provided scripts.

   Script for CPU profile: `./bin/cpu_profiling`

   Script for Memory profile: `./bin/memory_profiling`

   *Note: Ensure that these scripts are given the proper permissions to run*

1. Observe results in a PDF format by running:
   ```
   go tool pprof --pdf dist/<GOOS>/<GOARCH>/secretless-broker /var/path/to/cpu.pprof > file.pdf
   ```

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
1. Run `./bin/prefill_changelog $(cat VERSION)` to populate the [changelog](CHANGELOG.md) with
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
