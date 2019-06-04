## Overview

These are notes on implementation details relevant only to developers.

`config.v2` is an adapter package for parsing `v2` `secretless.yml` files and
converting them into `v1.Config` objects.   This approach let us overhaul the
user config experience, without updating the guts of Secretless.

## Design Details

The conversion occurs in 3 distinct steps, starting with the `yml` file input:

1. v2 `secretless.yml` file (input)
2. Raw `yml` struct (`v2.ConfigYAML`)
3. Programmer friendly representation (`v2.Config`)
4. Desired output (`v1.Config`)

Adding more detail:

1. A v2 `secretless.yml` is parsed into Go structs whose fields and structure
map 1-1 onto the fields and structure of the yml file.  You can think of these
as "literal struct versions of the raw yml file sections".  The top-level type
is `v2.ConfigYAML`, all these types have the suffix `YAML`,  and they are
defined in `config_yaml.go`.

2. The `xxxYAML` types are converted into programmer friendly types that aren't
constrained by the structure of `secretless.yml`.  The top-level type here is
`v2.Config`. You can think of this step as converting to "representations
convenient for code".  Eg, in the yaml, credentials are represented as a
dictionary whose keys are the credential names. This makes the yaml more compact
and readable.  However, once inside the code, we want a self-contained concept
of `Credential` which includes a field for the credential's name.

3. The real conversion logic: field name changes, structural transformations,
and the protocol specific transformations defined in
`v2.Service.ProtocolConfig`.

## Motivation for Design

`v1` used a messier approach for parsing, where fields whose only purpose was to
act as temporary receptacles for unmarshalled yaml leaked into the objects used
everywhere in the application. 

As a concrete example, `v1.Handler.Match` exists only because there is a field
named `match` in the `v1` yaml config files. The application itself is only
interested in `v1.Handler.Patterns`, which holds the regexes derived from the
strings in `Match`.  That is, `Match` is only an implementation artificat of
the parsing stage, but it has leaked into application code.

The approach we take in the `v2` parsing avoids the problem.  It also adds
clarity, since each stage has a specific, well-defined responsiblity.

## Additional Notes

In both the `v1` and `v2` cases, a Secretless configuration is, fundamentally,
just a list of proxy service definitions.  In `v2`, this is explicit, as you
can see from the definition of a `v2.Config`:

```go
type Config struct {
	Services []*Service
}
```

However, this concept is only implicit in `v1`, where a configuration is list
of handlers and list of listeners:

```go
type Config struct {
	Listeners []Listener
	Handlers  []Handler
}
```

In package `v2`, we introduce the concept of a `v1Service` (a
`Listener`/`Handler` pair) to make the concept explicit, and simplify the
translation of configuration from one package to the other.
