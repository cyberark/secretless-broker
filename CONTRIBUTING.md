# Contributing to the Secretless Broker

For general contribution and community guidelines, please see the [community repo](https://github.com/cyberark/community). In particular, before contributing
please review our [contributor licensing guide](https://github.com/cyberark/community/blob/main/CONTRIBUTING.md#when-the-repo-does-not-include-the-cla)
to ensure your contribution is compliant with our contributor license
agreements.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Pull Request Workflow](#pull-request-workflow)
- [Style Guide](#style-guide)
- [Building](#building)
- [Testing](#testing)
- [Documentation](#documentation)
- [Profiling](#profiling)
- [Plugins](#plugins)
- [Submodules](#submodules)
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

[design]: https://github.com/cyberark/secretless-broker/tree/main/design
[issues]: https://github.com/cyberark/secretless-broker/issues
[style]: STYLE.md
[tests]: #testing

## Building

First, clone `https://github.com/cyberark/secretless-broker` with the
`--recurse-submodules` flag. If you already have secretless-broker cloned locally,
but are missing submodules, perform `git submodule update --init --recursive`.
If you're new to Go, be aware that Go can be very selective about where the files
are placed on the filesystem. There is an environment variable called `GOPATH`,
whose default value is `~/go`. Secretless Broker uses
[go modules](https://golang.org/cmd/go/#hdr-Modules__module_versions__and_more)
which require either that you clone this repository outside of your `GOPATH` or
you set the `GO111MODULE` environment variable to `on`. We recommend cloning this
repository outside of your `GOPATH`.

Once you've cloned the repository, you can build the Secretless Broker.

Note: On git submodules, taken from git documentation.
> Luckily, you can tell Git (>=2.14) to always use the --recurse-submodules flag by setting the
> configuration option submodule.recurse: git config submodule.recurse true.
> As noted above, this will also make Git recurse into submodules for every
> command that has a --recurse-submodules option (except git clone)

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
2. Jenkins will automatically search the `test` directory for any sub-directories that
meet the criteria of having both a `start` and `stop` script. It then runs the
`./bin/run_integration` script on that directory.

Note: You can test locally using the same format of `./bin/run_integration <test
directory name>`. You can pass in the name of the directory itself, you don't need the
full path.

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

## Submodules

Secretless makes use of some third party libraries using Git Submodules.
In the [build instructions](#building) we cover some of the basics of how to
clone the repository and populate the submodules. In this section, we cover
how to work in the submodules of the Secretless project.

Development on submodules is similar to just working with a second repository, in that
you can `cd` into it and check out branches or make separate commits. However, you also
have the ability to commit and push recursively from the parent repository. For help
with this, it is recommended to review the "Publishing Submodule Changes" section of
the [Git Submodules](https://git-scm.com/book/en/v2/Git-Tools-Submodules) documentation.

`git submodule update --recursive` can be used to update the registered
 submodules to match what the superproject expects by cloning missing submodules and
 updating the working tree of the submodules. The "updating" can be done in several
 ways depending on command line options and the value of submodule.<name>.update
 configuration variable. The command line option takes precedence over the
 configuration variable. If neither is given, a checkout is performed. The recursive
 flag will handle any submodules nested within a submodule

### Updating Submodules
When making a change to a submodule, it will not be committed automatically when the
super repository is committed. As such, there are a few steps in place to make sure
this happens.

1. Enter the submodule directory and create a branch to store your changes, publishing
this branch now or once you have completed your work.

`cd third_party/<submodule>`

`git checkout -b <branch name>`

2. Commit any neccessary changes within the submodule repository to your new branch. If
you haven't pushed the branch, and the changes within, be sure to do so now with
`git push`. This is the same workflow, and end result, as if you were working on an
individual repository.

3. From the super directory, you'll notice that if you run `git status`, it will detect
new commits to the repository.
For example:
```
Changes not staged for commit:
(use "git add <file>..." to update what will be committed)
(use "git restore <file>..." to discard changes in working directory)
    modified:   third_party/go-mssqldb (new commits)
```

4. From the super directory, stage the changes to the submodule. It should only require
a single `git add <submodule_dir>` statement. This will update the remote that the
super repository uses to reference the submodule to point to the new branch you
created, and all the commits contained within.

5. Push your changes from the super repository. Be sure you are recursively checking
for changes to your submodule, so that you don't leave anything behind, with:
`git push --recurse-submodules=check`. Again, this check will be performed if you have
modified your git config to always recurse into submodules. When you push, the
working branch for the super repository will have the commit or branch for the
submodule tied to it, but no changes will be made to it beyond what you did yourself
in the submodule directory.

6. When you create a Pull Review in Github, you will notice that, within the 'files
changed' tab, there is a single reference to the changes made in the submodule, with
links to the github pages for them as well.

7. Create a seperate Pull Review for your submodule in its respective repository. This
is an extra measure to make sure both repositories are reviewed before their changes
are merged.

There are a few benefits to this approach
- When a pull request or branch is dependent on a specific commit in a submodule, we
 can easily pull both at the same time and build without issues.
- Reviewers can see the context for a change that may span more than one repository
 when Github links the two pull requests.

We want secretless to point to specific commits within a submodule, rather
than main. Make sure your change to secretless considers this.
 1. `cd` into the submodule
 2. Checkout the commit we want secretless to use from within the submodule
 3. Return to the secretless-broker directory, and create a new PR with the modified
    reference

 ### Helpful Commands for working with submodules
 To check the current hash for a submodule:

 `git ls-tree <branch> third_party/<submodule-directory>`

 To set (checkout) the SHA-1 of a submodule to the most recent commit:
 `git submodule --update <optional directory path>`

## Releasing

### Verify and update dependencies
1. Review the changes to `go.mod` since the last release and make any needed
   updates to [NOTICES.txt](./NOTICES.txt):
   - Add any dependencies that have been added since the last tag, including
     an entry for them alphabetically under the license type (make sure you
     check the license type for the version of the project we use) and a copy
     of the copyright later in the same file.
   - Update any dependencies whose versions have changed - there are usually at
     least two version entries that need to be modified, but if the license type
     of the dependency has also changed, then you will need to remove the old
     entries and add it as if it were a new dependency.
   - Remove any dependencies we no longer include.

   If no dependencies have changed, you can move on to the next step.

### Update the version and changelog
1. Create a new branch for the version bump.
1. Based on the unreleased content, determine the new version number and update
   the [version.go](pkg/secretless/version.go) file.
1. Review the [changelog](CHANGELOG.md) to make sure all relevant changes since
   the last release have been captured. You may find it helpful to look at the
   list of commits since the last release - you can find this by visiting the
   [releases page](https://github.com/cyberark/secretless-broker/releases) and
   clicking the "`N commits` to main since this release" link for the latest
   release.

   This is also a good time to make sure all entries conform to our
   [changelog guidelines](https://github.com/cyberark/community/blob/main/Conjur/CONTRIBUTING.md#changelog-guidelines).
1. Commit these changes - `Bump version to x.y.z` is an acceptable commit message - and open a PR
   for review. Your PR should include updates to `pkg/secretless/version.go`,
   `CHANGELOG.md`, and if there are any license updates, to `NOTICES.txt`.

### Add a git tag
1. Once your changes have been reviewed and merged into main, tag the version
   using `git tag -s v0.1.1`. Note this requires you to be  able to sign releases.
   Consult the [github documentation on signing commits](https://help.github.com/articles/signing-commits-with-gpg/)
   on how to set this up. `vx.y.z` is an acceptable tag message.
1. Push the tag: `git push vx.y.z` (or `git push origin vx.y.z` if you are working
   from your local machine).

### Create a GitHub pre-release
**Note:** Until the stable quality exercises have completed, the GitHub release
should be officially marked as a `pre-release` (eg "non-production ready")

1. From the Jenkins pipeline for the tag, retrieve the archived `dist/goreleaser`
   directory.
1. Create a GitHub release from the tag, add a description by copying the CHANGELOG entries
   from the version, and upload the release artifacts from `dist/goreleaser`
   to the GitHub release. The following artifacts should be uploaded to the release:
   - CHANGELOG.md
   - NOTICES.txt
   - LICENSE
   - secretless-broker_{VERSION}_amd64.deb
   - secretless-broker_{VERSION}_amd64.rpm
   - secretless-broker_{VERSION}_darwin_amd64.tar.gz
   - secretless-broker_{VERSION}_linux_amd64.tar.gz
   - SHA256SUMS.txt

   You should also locally rename the `secretless-broker` binaries in the
   `secretless-broker-{OS}_{GOOS}_{GOARCH}` dirs to `secretless-broker-{GOOS}`
   and upload these to the release.
1. Copy the `secretless-broker.rb` homebrew formula output by goreleaser
   to the [homebrew formula for Secretless](https://github.com/cyberark/homebrew-tools/blob/main/secretless-broker.rb)
   and submit a PR to update the version of Secretless available in brew.

### Publish the Red Hat image
1. Visit the [Red Hat project page](https://connect.redhat.com/project/3100131/view) once the images have
   been pushed and manually choose to publish the latest release.
