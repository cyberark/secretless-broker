## Generic Authentication - HTTP(S)

Secretless comes with built-in connectors that allow you to connect to services like AWS, or services that use Basic authentication.  But what if you want to connect to a service that Secretless has no connector for?  No problem!  The Generic HTTP connector can handle most services.  It gives you the ability to inject authentication types and their respective credentials directly into a header, as well as functions that can be used for encrypting the contents of your request.

--------------------------------
### Credentials

- pattern

    _Optional_

    Regular expressions can be used to validate their respective credential fields.

    Example:
    ~~~
    credentials:
        username: "foo"
        pattern: ^.+$
    ~~~

--------------------------------
### Config

- headers

    _Required_

    Here, you pass in your various HTTP headers, including your authorization. 
    Authorization values can be passed in as a single string, making use of credential fields and method calls 
    through string interpolation (i.e. double brace syntax).

    Replacement of template variables through method calls always occurs before functions are run.

    Example:
    ~~~
    credentials:
        username: "foo"
    config: 
        Authorization: "Basic {{ "bar" + username }}"
    ~~~

- forceSSL

    _Optional_

    A boolean value which modifies the HTTP scheme and forces connection over HTTPS if true.

    Example:
    ~~~
    configuration:
        headers
            ...
        forceSSL: true
    ~~~

--------------------------------

### Examples

This example authenticates using Basic authentication, with the username and password fields passed into a Base64 hash function.
~~~
services:
  http_good_basic_auth:
    connector: generic_http
    listenOn: tcp://0.0.0.0:8080
    credentials:
      username: 
        value: someuser
        pattern: ^[^:]+$
      password: testpassword
        value: somepassword
        pattern: ^.+$
    config:
      headers:
        Authorization: "Basic {{ Base64(username + ":" + password) }}"
      forceSSL: true
      match:
        - ^http
~~~
