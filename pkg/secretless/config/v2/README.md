##Overview

`config.v2` is an adapter package for parsing v2 "secretless.yml" files and
converting them into v1.Config objects.   This approach let us overhaul the user
config experience, without updating any of the guts of Secretless.

##Design Details

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
`v2.Service.ProtocolConfig`".

##Additional Notes

##Motivation for Design

v1 conversion used a messier approach where fields whose only purpose was
for the temp parsing leaked into the objects use in the code. eg, Match
