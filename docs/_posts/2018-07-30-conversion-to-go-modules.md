---
layout: post
title: "Converting Secretless to Golang modules"
date: 2018-07-30 09:00:00 -0600
author: Srdjan Grubor
categories: blog
published: true
excerpt: "How we converted Secretless to the new Golang dependency management"
---

## Introduction

There has been a lot of buzz lately about Go modules but the information out is a bit slim about what they are and how they fit into the future development of Go projects. Let us first take a look at a bit of history on how we got to where we are in terms of Go dependency management tooling.

## At first, there was nothing

If you ever developed in Go and worked with other programming language is, you might have noticed a different workflow when it came to dependencies. In Golang, all of your code for all your projects including their dependencies was assumed to be stored in your GOPATH location (usually `~/go`) which is rather unique way of checkpointing the state of your local dependencies. While at first look it is hard to understand why this is so, we have to understand a bit who created the language: Google.

### Golang at Google

I have to preface this by saying that my tenure with Go is short and I do not work for Google but it is widely known that Google keeps majority of its code in a [big monorepo](https://cacm.acm.org/magazines/2016/7/204032-why-google-stores-billions-of-lines-of-code-in-a-single-repository/fulltext). Since Go was developed there initially as an internal tool, the need for dependency management and storing your code in more than a single location would not have made sense in such a context. In other words, with a monorepo, each `git checkout` and `git commit` is tied exactly to the state of all dependencies you would want to care about so "dependency management system" outside of `git` itself and the rather-rudimentary `go get` command isn't really needed.

With the rise of Docker, Kubernetes, and other containerization tools written in Go, the wider community has picked up the language but unlike within Google, most developers:
- do not store all their code in a single repo
- do not want only a single version of a dependency shared between projects
- do not store all their dependencies for all projects in a single location

With these changes in use cases and developer profiles as Go was getting more traction outside of Google, it is not hard to understand why in the last couple of years dependency management has become such a hotly debated topic.

## Nature finds a way

Lack of serious push upstream for a real dependency management other than the [vendor directory](https://blog.golang.org/go1.5) seems to have motivated the wider community to make their own solution. Many projects have tried to fill this void with their [own visions](https://medium.com/@sdboyer/so-you-want-to-write-a-package-manager-4ae9c17d9527) of what this type of system should be (such as dep, Godep, Govendor, and many others). Within the last year or so, things have stabilized to some degree with [`dep`](https://github.com/golang/dep) emerging as a _de facto_  dependency management system and even being titled an "official experiment". Somewhere around this point, Google started working on creating an official dependency management system that would be integrated within the binary itself.

## `vgo` and the new Go modules

The result of upstream research into this culminated with a [vgo prosal](https://research.swtch.com/vgo-intro) and a sample implementation. These changes seemed to have garnered a lot of heated discussions in the community over the last few months but in the end `vgo` is now available as the built-in dependency manager. This feature has since been added in `golang:1.11beta2` ([docker image](https://hub.docker.com/_/golang/)) and should be available in the GA version of 1.11 around 1st of August (a day after this posting).

## Conversion of Secretless

At this point you might be asking yourself "why would we want to change from already-working system to the new modules"? The answers to this are numerous but the main ones are:
- Use of official upstream-supported tool
- Anticipated future deprecation of current non-official tooling
- Simpler builds
- Faster builds
- Directory-independent development

Also, we have more freedom to change things now, early in the development of Secretless which will become much harder to do as the project grows.

So with all of this in mind, let us see what is needed to convert your project to `vgo`-style one.

### Step #1 - Move your code out of `$GOPATH`

Because we won’t use the old “vendor” storage, we need to move our code away from `$GOPATH` to activate the automatic module processing. The current Golang (v1.11) way of activating modules follows two paths:
- If you use go modules outside of `$GOPATH`, you don’t need anything - the modules are used by default
- If you still use `$GOPATH`, modules are vendored by default unless you have `export GO111MODULE=on` set in your environment

Since dealing with environment variables is cumbersome and we want to use the same approach most other languages use, relocating the code somewhere outside of the `$GOPATH` is highly preferred.

### Step #2 - Convert your old tooling definitions to `go.mod`

This step is rather trivial and only takes a single command. Even if you don’t have `dep` as your dependency management tool, `go mod` is designed to handle conversion from most tools you might be using. If you’re running this inside a container, make sure that you have `git` installed beforehand.

_Note: Codebase no longer needs to reside in the GOPATH - in fact, `go mod` will warn you if you try to have it in the path as mentioned in the previous step!_

```
$ go mod -v -init -module github.com/conjurinc/secretless
go: creating new go.mod: module github.com/conjurinc/secretless
go: copying requirements from Gopkg.lock
```

You should now have a new `go.mod` file present in your repository. While `vendor/` directory is currently still supported on top of go modules, it is a deprecated way to store your dependencies moving forward and you may remove it.

_Note: If this conversion works, don't forget to commit the `go.mod` and remove your `Gopkg.*` files in your repository._

_Note #2: As we found out during conversion, if you have local modules that your code depends on you may need to manually add things to your `go.mod` files by:
- Adding the dependency name in the `require` section with a version number like this: `github.com/conjurinc/secretless v0.3.0`
- Adding a `replace` directive at the bottom of the file like this: `replace github.com/conjurinc/secretless => ../secretless`_

### Step #3 - Sync Your Dependencies

Now that we have our module file, it is time to get the dependencies downloaded with `go mod -sync` and ensure that we don’t have any junk in the `go.mod` file:
```
$ go mod -sync
go: finding golang.org/x/text v0.0.0-20171227012246-e19ae1496984
…
go: downloading golang.org/x/text v0.0.0-20171227012246-e19ae1496984
```

This stage will create `go.sum` which will contain verification hashes of modules that you `sync`d and remove unused imports from `go.mod` so make sure to commit these two files again or amend the previous commit.


### Step #4 - Run Your Code!

Well, it's as simple as that - you're done! There are no more steps to do other than running your codebase!

```
$ go run cmd/secretless/main.go  -f test/http_basic_auth/secretless.yml
2018/07/17 18:17:18 Secretless starting up...
```

_Note: If you did not run the sync command, `go run` and `go build` will fetch all the dependencies for you in this step so sync might not be strictly needed_

## Conclusion

While our conversion was not as trivial as many other blogs have led us to believe, we were able to do this in about 16 dev-hours for two codebases. The changeover has reduced bloat and made dependencies much simpler/faster and we are in good shape on this from for the forseeable future. Even though there are kinks to work out and there are some odd architecture choices made in `vgo` it is looking like a much needed step in the right direction for Golang tooling.

## Addendum

#### Adding dependencies

Adding dependencies works incredibly simple just by using `go get <path>`. It will automatically be added to both `go.mod` and `go.sum`.

#### Listing dependencies

To list all the modules listed as dependencies in `go.mod`, run the following: `go list -m all`.

#### Removing dependencies

Removing modules is a touch trickier but still pretty simple: `go mod -droprequire=<path>` or manually editing `go.mod`.

#### Docker dependency caching

If you built Go or Node modules within Docker and you do it often, you might know that caching the downloaded modules in a separate Docker layer can improve your build time by a large amount. Unlike with `dep`, current support for this is clunky now with the new modules ([GitHub issue](https://github.com/golang/go/issues/26610)). The current way to fetch all the dependencies with only `go.mod` and `go.sum` files is `go list -e $(go list -m all 2>/dev/null | awk '{print $1}’)` but keep an eye out on the issue linked as this may get fixed at some point.
