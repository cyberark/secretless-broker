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
- [Experience (Summary)](#experience-summary)
- [Experience (Detailed)](#experience-detailed)
- [Technical Details](#technical-details)
- [Testing](#testing)
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
  postgres-db:
    protocol: pg
    listenOn: 0.0.0.0:5432 # can be a socket as well (same name for both)
    credentials:
      address: postgres.my-service.internal:5432
      password:
        providerId: name-in-vault
        provider: vault
      username:
        providerId: username
        provider: env
    config:  # this section usually blank
      optionalStuff: blah
      
  # the aws prefix on the credentials indicates which protocol implementation to use
  aws-client:
    protocol: http
    listenOn: /var/docker/docker.sock
    credentials:
      aws/accessKeyID:
        providerId: name-in-vault
        provider: conjur
      aws/secretAccessKey:
        providerId: name-in-vault
        provider: conjur
      aws/accessToken:
        providerId: name-in-vault
        provider: conjur
    config:
      pattern: ^http.*

  # the conjur prefix on the credentials indicates which protocol implementation to use
  conjur-client:
    protocol: http
    listenOn: 127.0.0.1:8080
    credentials:
      conjur/accessToken:
        providerId: /path/to/file
        provider: file
      conjur/forceSSL:
        providerId: name-in-vault
        provider: conjur
    config:
      pattern: ^http://srdjan.com*

  ssh-handler:
    protocol: ssh
    listenOn: 0.0.0.0:2222
    credentials:
      address: "localhost"
      user: "Jonah"
      privateKey:
        providerId: SSH_PRIVATE_KEY
        provider: env
```
**Note**: We have not decided on the final syntax for http yet. We are considering dropping support in the near-term for
multiple http sub-handlers on a single port / socket file.
Some things we'd like to be able to support with the http definition:
- Connections to multiple different endpoints flowing through a single port / socket file
- Supporting connections to multiple different versions of the same connection type through the same port / socket file
  (eg connections to different AWS accounts)
- A clear definition for one listener / handler configured on a given port / socket (**this is the priority to support in this version**)

This proposal:
- Integrates the port / socket that Secretless is listening on with how to process the connection (eg credentials, etc)
- It is simple and easy to parse, with everything for a given connection in one small snippet of YAML
- It leverages keys as names to minimize the number of lines required to configure a given handler
- It enables separating configuration and credential data
- It will reduce the options available for configuring the http handler to simplify the experience
- It does not expose the concepts of listeners and handlers, but rather uses naming conventions that will do a better job
  of mapping to the user's mental model of what the broker is doing.

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
### Development
- #708 - Configuration parsing in config package handles new config format
- #714 - Secretless configuration syntax is versioned 

### Testing
- #712 - All Secretless test cases use new yml config format

### Documentation
- #713 - All Secretless configuration code samples are updated to reflect new yml config format
- cyberark/secretless-docs#135 - Listener / Handler references are updated to refer to single handler construct

### Future Work
- We have planned future work to further simplify and improve the plugin interface. This will be done in a follow-on effort,
  and will include updating the structure of the configuration objects. These changes to the configuration objects will be
  internal and not visible to end-users.
- Update the configuration CRD to follow the new model (#715) - until we make these updates, CRDs will remain in beta and
  will use the old style Secretless configuration
