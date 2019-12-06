# Table of Contents

- [Golang Style Guide and Best Practices](#golang-style-guide-and-best-practices)
  * [Style Baseline: Effective Go](#style-baseline--effective-go)
  * [Common Go Code Review Comments](#common-go-code-review-comments)
  * ["Line of Sight" Coding Best Practice](#-line-of-sight--coding-best-practice)
  * [Unit Testing Best Practices](#unit-testing-best-practices)
    + [Code Coverage](#code-coverage)
    + [Mocking](#mocking)
    + [Table Drive Testing](#table-drive-testing)
  * [Package Management](#package-management)
  * [Copyright Notices](#copyright-notices)
  * [Appendix](#appendix)
    + [References](#references)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

# Golang Style Guide and Best Practices

This document serves as a style and best practices guide for CyberArk developers who are writing Golang source code.

## Style Baseline: Effective Go 

The [Effective Go](https://golang.org/doc/effective_go.html) style guide serves as a baseline for our Golang style guidelines. Some handy links to main sections of this wiki document are provided here:

[Formatting](https://golang.org/doc/effective_go.html#formatting)

[Commentary](https://golang.org/doc/effective_go.html#commentary)

[Names](https://golang.org/doc/effective_go.html#names)

[Semicolons](https://golang.org/doc/effective_go.html#semicolons)

[Control structures](https://golang.org/doc/effective_go.html#control-structures)

[Functions](https://golang.org/doc/effective_go.html#functions)

[Data](https://golang.org/doc/effective_go.html#data)

[Initialization](https://golang.org/doc/effective_go.html#initialization)

[Methods](https://golang.org/doc/effective_go.html#methods)

[Interfaces and other types](https://golang.org/doc/effective_go.html#interfaces_and_types)

[The blank identifier](https://golang.org/doc/effective_go.html#blank)

[Embedding](https://golang.org/doc/effective_go.html#embedding)

[Concurrency](https://golang.org/doc/effective_go.html#concurrency)

[Errors](https://golang.org/doc/effective_go.html#errors)

## Common Go Code Review Comments

The [Go Code Review Comments wiki](https://github.com/golang/go/wiki/CodeReviewComments#go-code-review-comments) provides a collection of common comments made during Golang code reviews. This can be viewed as a supplement to the  [Effective Go](https://golang.org/doc/effective_go.html) style guide.

[Gofmt](https://github.com/golang/go/wiki/CodeReviewComments#gofmt)

[Comment Sentences](https://github.com/golang/go/wiki/CodeReviewComments#comment-sentences)

[Contexts](https://github.com/golang/go/wiki/CodeReviewComments#contexts)

[Copying](https://github.com/golang/go/wiki/CodeReviewComments#copying)

[Crypto Rand](https://github.com/golang/go/wiki/CodeReviewComments#crypto-rand)

[Declaring Empty Slices](https://github.com/golang/go/wiki/CodeReviewComments#declaring-empty-slices)

[Doc Comments](https://github.com/golang/go/wiki/CodeReviewComments#doc-comments)

[Don't Panic](https://github.com/golang/go/wiki/CodeReviewComments#dont-panic)

[Error Strings](https://github.com/golang/go/wiki/CodeReviewComments#error-strings)

[Examples](https://github.com/golang/go/wiki/CodeReviewComments#examples)

[Goroutine Lifetimes](https://github.com/golang/go/wiki/CodeReviewComments#goroutine-lifetimes)

[Handle Errors](https://github.com/golang/go/wiki/CodeReviewComments#handle-errors)

[Imports](https://github.com/golang/go/wiki/CodeReviewComments#imports)

[Import Blank](https://github.com/golang/go/wiki/CodeReviewComments#import-blank)

[Import Dot](https://github.com/golang/go/wiki/CodeReviewComments#import-dot)

[In-Band Errors](https://github.com/golang/go/wiki/CodeReviewComments#in-band-errors)

[Indent Error Flow](https://github.com/golang/go/wiki/CodeReviewComments#indent-error-flow)

[Initialisms](https://github.com/golang/go/wiki/CodeReviewComments#initialisms)

[Interfaces](https://github.com/golang/go/wiki/CodeReviewComments#interfaces)

[Line Length](https://github.com/golang/go/wiki/CodeReviewComments#line-length)

[Mixed Caps](https://github.com/golang/go/wiki/CodeReviewComments#mixed-caps)

[Named Result Parameters](https://github.com/golang/go/wiki/CodeReviewComments#named-result-parameters)

[Naked Returns](https://github.com/golang/go/wiki/CodeReviewComments#naked-returns)

[Package Comments](https://github.com/golang/go/wiki/CodeReviewComments#package-comments)

[Package Names](https://github.com/golang/go/wiki/CodeReviewComments#package-names)

[Pass Values](https://github.com/golang/go/wiki/CodeReviewComments#pass-values)

[Receiver Names](https://github.com/golang/go/wiki/CodeReviewComments#receiver-names)

[Receiver Type](https://github.com/golang/go/wiki/CodeReviewComments#receiver-type)

[Synchronous Functions](https://github.com/golang/go/wiki/CodeReviewComments#synchronous-functions)

[Useful Test Failures](https://github.com/golang/go/wiki/CodeReviewComments#useful-test-failures)

[Variable Names](https://github.com/golang/go/wiki/CodeReviewComments#variable-names)

## "Line of Sight" Coding Best Practice

References:

[Indent Error Flow](https://github.com/golang/go/wiki/CodeReviewComments#indent-error-flow)

[Things in Go I Never Use](https://www.youtube.com/watch?v=5DVV36uqQ4E&feature=youtu.be&t=660) video by Mat Ryer

## Unit Testing Best Practices

### Code Coverage

### Mocking

TODO: Need a description of how to propagate mockable interfaces from top level, so that interfaces can be mocked at each level of unit testing.

### Table Drive Testing

Reference: [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests) wiki

## Package Management

References:

[Using Go Modules](https://files.slack.com/files-pri/T0A233U2Z-FR043S136/download/local-dev-go-mod.mp4) video by Kumbirai Tanekha

[Go Modules](https://github.com/golang/go/wiki/Modules) wiki

[Using Go Modules](https://blog.golang.org/using-go-modules)

## Copyright Notices

## Appendix

### References

[Golang language specification](https://golang.org/ref/spec)

[Tour of Go](https://tour.golang.org/)

[How to Write Go Code](https://golang.org/doc/code.html)

[Effective Go](https://golang.org/doc/effective_go.html) Style Guide

[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments#go-code-review-comments)

[Common Go Coding Mistakes](https://github.com/golang/go/wiki/CommonMistakes)

[Awesome Go](https://awesome-go.com/): A curated list of awesome Go frameworks, libraries and software.

Video: [Go Best Practices](https://www.youtube.com/watch?v=MzTcsI6tn-0&feature=youtu.be&t=1457) by Ashley McNamara & Brian Ketelsen

Video: [Things in Go I Never Use](https://www.youtube.com/watch?v=5DVV36uqQ4E&feature=youtu.be&t=660) by Mat Ryer (Good overview of "Line-of-Sight" coding)

[Context Use and Misuse](https://peter.bourgon.org/go-for-industrial-programming/#context-use-and-misuse) by Peter Bourgon

[Golang Articles](https://github.com/golang/go/wiki/Articles)

[Golang `template` Package](https://golang.org/pkg/text/template/)

