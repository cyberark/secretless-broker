/*

Package v2 is a package for parsing version 2 secretless.yml files.  Most users of this package will only be concerned
with the single func NewConfig, which parses yaml file content.

File Format

Here is an example configuration for an http basic auth service that
demonstrates all the features of a v2 yaml file:

    version: 2
    services:
      http_basic_auth:
        connector: basic_auth
        listenOn: tcp://0.0.0.0:8080
        credentials:
          username: someuser
          password:
            from: conjur
            get: testpassword
          config:
            authenticateURLsMatching:
              - ^http.



A few notes:

    - listenOn:
        This may be a tcp port on localhost or a unix socket.  tcp ports should
        start with tcp:// and sockets with unix://.  A socket address might look
        like: unix:///some/absolute/path.

    - credentials:
        The keys of this dictionary are the names of the credentials within
        secretless.  All values must be either a constant string, or a
        dictionary with the keys "from" and "get".  Dictionary keys specify the
        location of the secret within a Provider, such as a vault or the system
        environment. "from" identifies the type of secret Provider, and "get" is
        the id of the secret within that Provider.

    - config:
        The config key provides optional, protocol-specific configuration
        options.  For many protocols, it can be omitted.  In the case http,
        however, we must specify both the type of http authentication (in our
        example, "basic_auth") as well as which requests should be authenticated
        (in our example, all of them).




*/
package v2
