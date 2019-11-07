# Secretless Broker Style Guide

Use this guide to maintain consistent style across the Secretless Broker project.

Generally, when you are contributing code changes to this project you should
leverage [golint][golint] as you work to get suggestions for style fixes you
should make.

To use `golint` on any project path, run
```
go get -u golang.org/x/lint/golint
golint ./[path]/...
```

## Other general guidelines

### Be consistent

> If you’re editing code, take a few minutes to look at the code
> around you and determine its style. If they use spaces around all
> their arithmetic operators, you should too. If their comments have
> little boxes of hash marks around them, make your comments have
> little boxes of hash marks around them too.

> The point of having style guidelines is to have a common vocabulary
> of coding so people can concentrate on what you’re saying rather
> than on how you’re saying it. We present global style rules here so
> people know the vocabulary, but local style is also important. If
> code you add to a file looks drastically different from the existing
> code around it, it throws readers out of their rhythm when they go
> to read it. Avoid this.

### Prose

Secretless Broker prose should be warm, direct, and clear about its intended
audience. When editing copy, follow the [Chicago Manual of
Style][cmos].

### Documentation

Documentation is written to be familiar with the reader. Use you & your to refer
to the reader, we us & our to refer to the contributors.

When giving examples, use a generic example that the reader can relate
to. For example: a website connecting to a database.

Give names to machines according to their function, eg
`database-server`. Give names to humans using the [alphabetical
convention][names] common in cryptography research.

### Tests

When writing tests, treat your descriptions as prose to be read and understood
by humans.

### Code

Secretless Broker code favors readability over cleverness.

### Continuous integration

Consider the guidelines in [The Ten-Factor CI Job][10factor].

[cmos]: http://www.chicagomanualofstyle.org/home.html
[names]: https://en.wikipedia.org/wiki/Alice_and_Bob#Cast_of_characters
[10factor]: http://tenfactorci-conjur.herokuapp.com
[golint]: https://github.com/golang/lint

## Contributing

If this guide is lacking, please feel free to open a pull request to
include additional guidelines.

Good guideline are:

* Distinctive

  The reader can tell which part of the guide covers the project
  they're working on.

* Definitive

  The readers can distinguish whether their work follows the
  guidelines or not.

* Actionable

  The reader has clear guidance on how to improve content that doesn't
  follow the style.
