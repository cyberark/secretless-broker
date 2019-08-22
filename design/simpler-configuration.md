# Secretless configuration is intuitive and simple
The Secretless configuration has been the same since the start of the project, but since the start of the project
we have learned more about the internal concepts, how they link together, and how to name them so that they are clear
to end-users and people who are learning about the project for the first time.

In this design proposal, we suggest an alternate Secretless configuration syntax that will make it simpler to configure
a Secretless Broker sidecar. We propose making these updates in a way that we will continue to support the old configuration
syntax at a low cost, while improving the test coverage for the configuration package and clarifying / simplifying the
project documentation.

In sum, these proposed changes are for simplifying the user-facing experience of configuring Secretless. They remove
unnecessary concepts and present better chrome to the user.  This improvement comes in two main flavors:

1. Shorter, simpler `secretless.yml` configuration files
2. Shorter, simpler "concept" documentation

Aha Card: https://cyberark.aha.io/features/AAM-<>

- [Objective](#objective)
- [Team](#team)
- [SDLC Timeline](#sdlc-timeline)
- [Experience](#experience)
- [Technical Details](#technical-details)
- [Testing](#testing)
- [Dependent Components](#dependent-components)
- [Documentation](#documentation)
- [Open Questions](#open-questions)
- [Stories](#stories)
- [Future Work](#future-work)

### Objective
- Make it simpler to configure Secretless Broker, including using language in the configuration spec that is more intuitive
- Improve the Secretless documentation by simplifying the concepts one needs to understand in order to make sense of how
  the project works
- Stabilize the configuration package so that end users can expect no near-term changes will need to be made to the
  configuration beyond the next stable release
- Ensure that CRDs using the legacy configuration may still be used with the project once the standard configuration
  syntax has been updated
  
### Team
- Engineering Manager: Geri Jennings (@izgeri)
- Engineers:
    - Jonah Goldstein (@jonahx)
    - Kumbirai Tanekha (@doodlesbykumbi)
    - Srdjan Grubor (@sgnn7)

### SDLC Timeline
|Stage|Status|ETA|Artifact|
|-|-|-|-|
|High Level Feature Doc|n/a|n/a|n/a|
|Functional Sign-off|n/a|n/a|n/a|
|Detailed Feature Doc|Pending|2019-05-23|[Link](https://github.com/cyberark/secretless-broker/pull/716)|
|Solution Sign-off|Pending|2019-05-23||
|Epic|Pending|2019-05-23|[Link](https://github.com/cyberark/secretless-broker/issues/707)|
|Execution|Pending|2019-06-13||

### Experience 
#### Current Experience
At current users can provide configuration using a file via a ConfigMap or using our configuration custom resource
definition. Either way the configuration follows a similar syntax:

```yaml
listeners:
  - name: test-app-pg-listener
    protocol: pg
    address: 0.0.0.0:5432
  - name: http_default
    protocol: http
    address: 0.0.0.0:80

handlers:
  - name: test-app-pg-handler
    listener: test-app-pg-listener
    credentials:
      - name: address
        provider: conjur
        id: test-secretless-app-db/url
      - name: username
        provider: conjur
        id: test-secretless-app-db/username
      - name: password
        provider: conjur
        id: test-secretless-app-db/password
      - name: sslmode
        provider: literal
        id: require
  - name: http_good_basic_auth_handler
    type: basic_auth
    listener: http_default
    match:
      - ^http.*
    credentials:
      - name: username
        provider: literal
        id: someuser
      - name: password
        provider: literal
        id: testpassword
  - name: aws
    listener: http_default
    match:
      - ".*"
    credentials:
      - name: accessKeyId
        value:
          environment: AWS_ACCESS_KEY_ID
      - name: secretAccessKey
        value:
        environment: AWS_SECRET_ACCESS_KEY
```

Some of the key concerns we have with this configuration are:
- It separates listeners and handlers conceptually, though we are moving toward combining these into a single `handler`
  concept.
- Having listeners and handlers separate in the configuration (where you have a `listeners` section and a `handlers`
  section) also makes it harder to quickly parse the document to see what processing will be done on connections that come
  in over a certain port or socket file. For example, the handler definition needs to refer to the listener name that it is
  linked with.
- The configuration is overly verbose:
  - Both handlers and listeners must be named
  - Rather than using keys as names, we require specifying the `name` of many things
- The current configuration conflates credentials and configuration data
- The http handler configuration is inconsistent over whether the name must match the http handler type or whether (as
  is consistent with other handlers) you can use a random name and specify the `type`.
- In expecting users to understand listeners and handlers to write a configuration, we're exposing an implementation detail
  instead of crafting the configuration in terms that will be clear and sensible to the user.
- In the current experience, we never took into account the user's mental model and designed our experience around that


#### Proposed Experience
```yaml
version: "1"
services:

  ###
  # database handler example
  ###

  postgres-db:
    connector: pg
    listenOn: tcp://0.0.0.0:5432 # can be a socket as well (same name for both)
    credentials:
      host: postgres.my-service.internal
      password:
        from: vault
        get: id-of-secret-in-vault
      username:
        from: env
        get: username
    config:  # this section usually blank
      optionalStuff: blah
      
  ###
  # http handler example
  ###
      
  # the config for the http protocol has one required value:
  #   `authenticateURLsMatching` which gives a regex pattern for request URIs that use Secretless for auth (eg `match`)
  
  aws-client:
    connector: http
    listenOn: unix:///var/docker/docker.sock
    credentials:
      accessKeyID:
        from: conjur
        get: id-of-secret-in-conjur
      secretAccessKey:
        from: conjur
        get: id-of-secret-in-conjur
      accessToken:
        from: conjur
        get: id-of-secret-in-conjur
    config:
      authenticateURLsMatching: ^http.*

  conjur-client:
    connector: http
    listenOn: http://127.0.0.1:8080
    credentials:
      accessToken:
        from: file
        get: /path/to/file
      forceSSL: true
    config:
      authenticateURLsMatching: ^http://srdjan.com*

  ###
  # ssh handler example
  ###

  ssh-proxy:
    connector: ssh
    listenOn: tcp://0.0.0.0:2222
    credentials:
      address: "localhost"
      user: "Jonah"
      privateKey:
        from: env
        get: SSH_PRIVATE_KEY
```
**Note**: In this iteration we are **not** supporting requests for multiple
http handler types flowing through a single service definition. We plan to add
support for this in a later iteration. For now, we only support one http
handler configured on a given port / socket.

This proposal:
- Integrates the port / socket that Secretless is listening on with how to process the connection (eg credentials, etc)
- It is simple and easy to parse, with everything for a given connection in one small snippet of YAML
- It leverages keys as names to minimize the number of lines required to configure a given handler
- It enables separating configuration and credential data
- It will reduce the options available for configuring the http handler to simplify the experience
- It does not expose the concepts of listeners and handlers, but rather uses naming conventions that will do a better job
  of mapping to the user's mental model of what the broker is doing.

One other difference is that we enable less verbose options for specifying
credentials. For almost all credential providers (Conjur, Kubernetes Secrets,
HashiCorp Vault, environment, file, keychain) the syntax is:
```yaml
credentials:
  secretName:
    from: credentialProvider
    get: id-of-secret-in-provider
```
In the example snippet above, `secretName` is the credential's name within
Secretless, and as such must match a key in the handler configuration (eg
`address` for SSH). `credentialProvider` must match the unique identifier of
the desired credential provider (eg `kubernetes` for Kubernetes Secrets).
`id-of-secret-in-provider` is the fully qualified ID of the secret in the
specified credential store.

The one exception to this syntax is the `literal` provider, which no longer
ever needs to be referenced by its `credentialProvider` identifier but instead
is invoked by simply providing the `secretName` and the string value that the
key should be set to:
```yaml
credentials:
  secretName: "my-secret-value"
```

A final difference is that we don't need to specify sockets and local addresses
using separate keys.  Both are valid values for a single `listenOn` key.  You
use the `unix://` prefix to specify a socket, and a `tcp://` or `http://` (both
are equivalent) prefix to specify a local TCP address.

### Technical Details
- The [configuration parser](https://github.com/cyberark/secretless-broker/tree/master/pkg/secretless/config) needs to be
  updated to parse the updated YAML, while maintaining support for the old YAML syntax.
  - This will be done by creating a new parser that will convert the new config into the old objects, thus mapping it into
    the existing code that uses the current configuration definition. We are not planning at this time to refactor the
    existing structure of the configuration objects.

### Dependent Components
1. [Sidecar Injector](https://github.com/cyberark/sidecar-injector)
1. [Configuration CRD](https://github.com/cyberark/secretless-broker/tree/master/internal/app/secretless/configurationmanagers/kubernetes/crd)
1. [File Configuration Watcher](https://github.com/cyberark/secretless-broker/blob/master/internal/app/secretless/configurationmanagers/configfile/fs_watcher.go)
1. [Demos](https://github.com/cyberark/secretless-broker/tree/master/demos/)
1. Documentation and tutorials
1. Architecture diagrams
1. All existing integration and end-to-end tests

All of these will need to be reviewed and updated as appropriate to use the new syntax and nomenclature.

### Testing
The unit test coverage for the configuration package will be improved from the 74% coverage at current as part of this effort. In particular, we will add new unit tests to ensure we maintain support for the old syntax as well as adding unit tests to validate support for the new syntax.

In addition, all integration and end-to-end tests will be updated to use the new syntax after validating that the test suite still passes with the old syntax once the code changes have been made.

### Open Questions
- How should Secretless fail on invalid configuration when watching?
- Should we integrate versioning as part of these changes?

### Stories
#### Development
- #708 - Configuration parsing in config package handles new config format
- #714 - Secretless configuration syntax is versioned 

#### Testing
- #712 - All Secretless test cases use new yml config format

#### Documentation
- #713 - All Secretless configuration code samples are updated to reflect new yml config format
- cyberark/secretless-docs#135 - Listener / Handler references are updated to refer to single handler construct

### Future Work
- We have planned future work to further simplify and improve the plugin interface. This will be done in a follow-on effort,
  and will include updating the structure of the configuration objects. These changes to the configuration objects will be
  internal and not visible to end-users.
- Update the configuration CRD to follow the new model (#715) - until we make these updates, CRDs will remain in beta and
  will use the old style Secretless configuration
- Multiple http handlers on a single port / socket
  We have not decided on the final syntax for http with multiple configured backends yet. Some things we'd like to be able to support with the http definition:
  - Connections to multiple different endpoints flowing through a single port / socket file
  - Supporting connections to multiple different versions of the same connection type through the same port / socket file
    (eg connections to different AWS accounts)
