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
3. Add the `implementing` label to the issue as you begin to work on it
4. Run tests as described [here][tests], ensuring they pass
5. Submit a pull request, linking the issue in the description (e.g. `Connected to #123`)
6. Add the `implemented` label to the issue, and ask another contributor to review and merge your code

In addition to technical workflow descriptions available in GitHub,
some of the project's technical design documents can be found in the project [design][design] folder.

## Style guide

Use [this guide][style] to maintain consistent style across the Secretless Broker project.

[design]: https://github.com/cyberark/secretless-broker/tree/master/design
[issues]: https://github.com/cyberark/secretless-broker/issues
[style]: STYLE.md
[tests]: #testing

## Building

First, clone `https://github.com/cyberark/secretless-broker`. If you're new to Go, be aware that Go can be very selective
about where the files are placed on the filesystem. There is an environment variable called `GOPATH`, whose default value
is `~/go`. Secretless Broker uses [go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more) which
require either that you clone this repository outside of your `GOPATH` or you set the `GO111MODULE` environment variable to
`on`. We recommend cloning this repository outside of your `GOPATH`.

Once you've cloned the repository, you can build the Secretless Broker.

### Static long version tags

In most of our build scripts we provide a static (compile-time) version augmentation so that
the final artifacts include the Git short-hash of the code used to build it so that it looks
similar to: `<sem_ver>-<git_short_hash>`. We do this in most cases by over-riding the `Tag`
variable value in `pkg/secretless` package with ldflags in this manner:
```
...
-ldflags="-X github.com/cyberark/secretless-broker/pkg/secretless.Tag=<git_short_hash>"
...
```

If you would like the same behavior and something other than the default `dev` tag, you will
need to add the same ldflags to your build commands or rely on the current build scripts to
create your final deliverable.

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

### Prerequisites

* **Docker** You need Docker to run the tests.

Build the project by running:

```sh-session
$ ./bin/build
```

Then run the test cases:

```sh-session
$ ./bin/test
```
### Adding New Integration tests

Each integration test exists in its own subdirectory within the "test"
directory.  The test runner assumes it has the following executable scripts,
which will be run in order:

- `./start` (required) - Performs setup work -- eg, spinning up test
   containers, populating a database with test data, etc.
- `./test` (required) - Runs the actual tests.  It is assumed to produce go
   test output on stdout.
- `./stop` (optional) - Performs cleanup work.

To add a new integration test, complete the following two steps:

1. Create a folder with test scripts as described above.
1. Add a new entry to the `Jenkinsfile` to exercise those test scripts using
   the `run_integration` script. In most cases, you will also call `junit` on
   the xml file that `run_integration` outputs in your test's subdirectory.

Here's an example `Jenkinsfile` entry:

```
stage('Integration: PG Handler') {
  steps {
    sh './bin/run_integration pg_handler'
    junit 'test/pg_handler/junit.xml'
  }
}
```

### OSX Keychain provider Test

**OSX Keychain provider**

If you are on a Mac, you may also test the OSX Keychain provider:
```sh-session
cd test/providers/keychain/
./start
./test
```
This test will not be run as part of the test suite, since it requires access
to the Mac OSX Keychain. You will be prompted for your password when running
this test, as it temporarily adds a generic password to your account, and
verifies that it can retrieve the value.

### Kubernetes CRD loading test

```sh-session
cd test/manual/k8s_crds
./deploy
```
This test currently does not run as part of the test suite.

**Code Climate**

We use Code Climate in our CI pipeline to perform linting and other style
checks.  The specific engines we use and their configuration is in
`.codeclimate.yml`.

To run linting checks via the Code Climate golint engine, simply run:

```sh-session
./bin/check_style
```

## Documentation
Secretless has a few sources for documentation: a website, a documentation subdomain, and godocs.

### Website
The [website](https://secretless.io) source is in the [docs](docs/) folder in this repository. It is generated using Jekyll.

The source includes:
- the website main page
- some old pages that redirect to the documentation subdomain
- tutorials
- godocs for the plugin API
- Secretless blog
- community info page
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
[Godocs](https://godoc.org/github.com/cyberark/secretless-broker) are auto-published, and our `./bin/build_website` script also generates godocs for our plugin API that are published to our website.

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
   version: "2"
   services:
     pg_tcp:
       connector: pg
       listenOn: tcp://0.0.0.0:15432
       credentials:
         host:
           from: env
           get: PG_HOST
         username: test
         password:
           from: env
           get: PG_PASSWORD
    ```

1. The type of profiling is explicitly defined in the initial command that runs Secretless. Run Secretless with the profile desired like so:
   ```
   $ PG_HOST=localhost \
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

Secretless supports using [Go plugins](https://golang.org/pkg/plugin/) to extend
its functionality. To learn about writing new Secretless plugins and for more
information on the types of plugins we currently support, visit the [plugin API directory](pkg/secretless/plugin).

## Releasing

### Verify and update dependencies
1. Check whether any dependencies have been changed since the last release by running
   `./bin/check_dependencies`. The script will tell you what has changed. Beware - the script at current DOES NOT appropriately handle `replace` directives - you will need to process these manually.

1. If any dependencies have changed, for each changed dependency in assets/license_finder.txt you'll need to do the following:

   - if it is a new dependency, add an approval to the dependency decisions fileusing the [LicenseFinder](https://github.com/Pivotal/LicenseFinder):
     ```
     docker run --rm \
       -v $PWD:/scan \
       licensefinder/license_finder \
       /bin/bash -lc "
         cd /scan && \
         license_finder approvals add \
           --decisions-file=assets/dependency_decisions.yml \
           [DEPENDENCY] --version=[VERSION]"
     ```

   - update the [spreadsheet](https://cyberark365.sharepoint.com/:x:/s/Conjur/Edko_eT7CfpEuPxnnbIEfmAB4j2ybNozY9B8QAIDOxKynQ?e=CfP6ym) with the updated dependency info (add / edit / remove a row), including a link to the relevant license file

   - prepare the revised NOTICES.txt by adding / removing / editing dependency
     information

   If no dependencies have changed, you can move on to the next step.

### Update the version and changelog
1. Create a new branch for the version bump.
1. Based on the unreleased content, determine the new version number and update
   the [version.go](pkg/secretless/version.go) file.
1. Run `./bin/prefill_changelog` to populate the [changelog](CHANGELOG.md) with
   the changes included in the release.
1. Commit these changes - `Bump version to x.y.z` is an acceptable commit message - and open a PR
   for review.

### Add a git tag
1. Once your changes have been reviewed and merged into master, tag the version
   using `git tag -s v0.1.1`. Note this requires you to be  able to sign releases.
   Consult the [github documentation on signing commits](https://help.github.com/articles/signing-commits-with-gpg/)
   on how to set this up. `vx.y.z` is an acceptable tag message.
1. Push the tag: `git push vx.y.z` (or `git push origin vx.y.z` if you are working
   from your local machine).

### Build a release
**Note:** Until the stable quality exercises have completed, the GitHub release
should be officially marked as a `pre-release` (eg "non-production ready")
1. From a **clean checkout of master** run `./bin/build_release` to generate
   the release artifacts.
1. Create a GitHub release from the tag, add a description by copying the CHANGELOG entries
   from the version, and upload the release artifacts from `dist/goreleaser`
   to the GitHub release. The following artifacts should be uploaded to the release:
   - CHANGELOG.md
   - NOTICES.txt
   - secretless-broker_{VERSION}_amd64.deb
   - secretless-broker_{VERSION}_amd64.rpm
   - secretless-broker_{VERSION}_linux_amd64.tar.gz
   - SHA256SUMS.txt
