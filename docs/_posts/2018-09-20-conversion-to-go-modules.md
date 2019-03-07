---
layout: post
title: "Converting the Secretless Broker to Golang modules"
date: 2018-09-20 09:00:00 -0600
author: Srdjan Grubor
categories: blog
published: true
image: secretless_logo_blog.jpg
thumb: secretless_logo_blog.jpg
image-alt: Secretless logo
excerpt: "How we converted the Secretless Broker to use the new Golang dependency management"
---

## Introduction

There has been a lot of buzz lately about Go modules, but there is still not much information available about what they are and how they fit into the future development of Go projects. Based on the information available, however, we recently updated the Secretless Broker to use Go modules for our dependency management. In this post, we will talk about what led us to make this decision, and some of the technical details of how we implemented this change. But first, let's look at a bit of history on how we got to where we are in terms of Go dependency management tooling.

## At first, there was nothing

If you write code in Go after working with other programming languages, you may notice a different workflow when it comes to dependencies. In the early stages of Go, all of your code for all your projects _including their dependencies_ were assumed to be stored in your `GOPATH` location (usually `~/go`), which is rather unique way of checkpointing the state of your local dependencies. It's reasonable to wonder why Golang didn't choose to follow well-established patterns that other modern programming languages have adopted; to understand why this might be so, it helps to consider the origins of the language at Google.

### Golang at Google

Go was [developed at Google initially](https://talks.golang.org/2012/splash.article) to make it easier for its software engineers to write modern fast code at Google's scale. In the early phases of the project, dependency versioning or vendoring was not supported; this may be due in part to the fact that (as is widely known) Google keeps the majority of its code in a [big monorepo](https://cacm.acm.org/magazines/2016/7/204032-why-google-stores-billions-of-lines-of-code-in-a-single-repository/fulltext). In that context, support for managing dependency versions or storing code in more than a single location would not have made much sense. In other words - with a monorepo, each `git checkout` and `git commit` is tied exactly to the state of all the dependencies that matter, so a "dependency management system" outside of `git` itself and the `go get` command isn't really needed.

## Nature abhors a vacuum

Since `Go1` was officially released in 2012, [interest in Go has grown](https://blog.golang.org/8years) over time. As adoption increases in the wider community, there is greater demand for dependency management that supports developers who:
- do not store all their code in a single repo
- do not want only a single version of a dependency shared between projects
- do not store all their dependencies for all projects in a single location

Since there has been no one official solution that addresses all of these issues, a number of [alternate solutions](https://hackernoon.com/the-state-of-go-dependency-management-6cc5f82a4bfa) have emerged over time. At the time we started the Secretless Broker project the best solution available was [`dep`](https://github.com/golang/dep) (released in mid-2017), which was known as an "official experiment" and had emerged as the _de facto_  dependency management system for Go. `Dep` was easy to use, and running `dep ensure` would update the vendor directory (officially supported in Go as of [`Go1.5`](https://blog.golang.org/go1.5)).

## `vgo` and the new Go modules

Despite the growing acceptance of `dep` as _the_ dependency management tool for Golang, in mid-2018 a [new proposal](https://blog.golang.org/versioning-proposal) emerged for managing Golang project dependencies, known as `vgo` or Versioned Go Modules. The proposal was accepted, and as of [`Go v1.11`](https://tip.golang.org/doc/go1.11#modules) modules are available as an alternative to `GOPATH`, with integrated support for versioning and package distribution.

## Conversion of the Secretless Broker

As mentioned earlier in the post, the Secretless Broker was built using the `dep` dependency manager. So why would we change at this point? Our primary justifications for switching from `dep` to Go modules are:
- To use the official upstream-supported tool
- In anticipation of the future deprecation of current non-official tooling (i.e. `dep`)
- Simpler builds
- Faster builds
- Directory-independent development

In addition, our project is in its early stages, which means we have more freedom to change things that may become harder to change as the project grows.

So with all of this in mind, let us see what is needed to convert your project to `vgo`-style dependency management.

### Step #1 - Move your code out of `$GOPATH`

```
mv $GOPATH/path/to/my-code /path/to/new/location/
```

Because we won’t use the old “vendor” storage, we need to move our code away from `$GOPATH` to activate the automatic module processing. The current Golang (v1.11) way of activating modules follows two paths:
- If you use go modules outside of `$GOPATH`, you don’t need anything - the modules are used by default
- If you still use `$GOPATH`, modules are vendored by default unless you have `export GO111MODULE=on` set in your environment

Since dealing with environment variables is cumbersome and we want to use the same approach most other languages use, relocating the code somewhere outside of the `$GOPATH` is highly preferred.

### Step #2 - Convert your old tooling definitions to `go.mod`

The instructions below should work even if you don't currently have `dep` as your current dependency management tool, since `go mod` is designed to handle conversion from most tools you might be using. Since we're converting from `dep`, though, the instructions include references to `dep`-specific artifacts  like `Gopkg.lock`.

To convert to using Go modules, you run the following command from your project root:

```
$ go mod init github.com/cyberark/secretless-broker
go: creating new go.mod: module github.com/cyberark/secretless-broker
go: copying requirements from Gopkg.lock
```

After running this command, you should have a new `go.mod` file in your repository. Once this file has been created, you can commit it and remove the `Gopkg.*` files from your repository.

You may also want to remove the `vendor/` directory; though it is currently still supported, it is expected that it [will be deprecated](https://github.com/golang/proposal/blob/master/design/24301-versioned-go.md#proposal) going forward.

#### Things to watch out for
- If you’re running this inside a container, make sure that you have `git` (and possibly Mercurial depending on your dependencies) installed beforehand.
- If you skipped step one and your code still resides in the `GOPATH`, `go mod` will give you a warning when you run this step!
- If you have local modules that your code depends on, you may need to manually add things to your `go.mod` file.

  For example, if you have two projects in the same directory, and `project-a` depends on `project-b`, you can manually update the `go.mod` file for `project-a` to include the versioned dependency on `project-b` in the `require` section with a `replace` directive at the bottom of the `go.mod` file:
  ```
  github.com/org/project-b v0.3.0
  ...
  replace github.com/org/project-b => ../project-b
  ```

  This ensures that `project-a` includes the local code from `project-b`.


### Step #3 - Sync Your Dependencies

Now that we have our module file, it is time to get the dependencies cleaned up and downloaded locally:
```
$ go mod tidy
go: finding github.com/conjurinc/secretless/internal/app/secretless/providers latest
…

$ go mod download
go: finding golang.org/x/text v0.0.0-20171227012246-e19ae1496984
…
go: downloading golang.org/x/text v0.0.0-20171227012246-e19ae1496984
```

This stage will create `go.sum`, which will contain verification hashes of modules that you synced. `go.mod` and `go.sum` files are likely to change, so make sure to commit these two files again or amend the previous commit.


### Step #4 - Run Your Code!

Well, it's as simple as that - you're done! There are no more steps to do other than running your codebase!

```
$ go run cmd/secretless/main.go  -f test/http_basic_auth/secretless.yml
2018/07/17 18:17:18 Secretless starting up...
```

_Note: If you did not run the sync command, `go run` and `go build` will fetch all the dependencies for you in this step so sync might not be strictly needed_

## Conclusion

While our conversion was not as trivial as many other blogs have led us to believe, we were able to do this in about 16 dev-hours for two codebases. The changeover has reduced bloat and made dependencies much simpler/faster, and we are in good shape on this for the forseeable future. The project is in its early stages, so there are still issues to resolve and ways to make the process even smoother, but in general `vgo` is looking like a much needed step in the right direction for Golang tooling.

## Addendum

#### Adding dependencies

Adding dependencies works incredibly simple just by using `go get <path>`. It will automatically be added to both `go.mod` and `go.sum`.

#### Listing dependencies

To list all the modules listed as dependencies in `go.mod`, run the following: `go mod graph`.

#### Removing dependencies

Removing modules is a touch trickier but still pretty simple: `go get <path>@none` or manually editing `go.mod`.

#### Docker dependency caching

If you built Go or Node modules within Docker and you do it often, you might know that caching the downloaded modules in a separate Docker layer can improve your build time by a large amount. Just like with `dep`, you copy the relevant files (`go.mod` and `go.sum`) and run a simple command: `go mod download`.
