# Using templates to create Secretless Connector Plugins

We created the templates in this directory to make it easier to add new
connectors to Secretless.

Before using the templates, be sure to read the [Secretless Connector Plugins
README](https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md)

To create a new Secretless connector plugin, follow these instructions:

1. Copy the relevant template directory (HTTP/TCP) into a folder on your local
   machine (or to `internal/plugin/connectors/<connector_type>` if you are building
   an internal connector).

   If you're not sure which connector type is appropriate for your target service,
   please refer to the [connector technical overview](https://github.com/cyberark/secretless-broker/tree/master/pkg/secretless/plugin/connector#technical-overview) for guidelines.

1. Update the copied files to implement your connector. Each file includes
   instructions in the form of `TODO`s.

1.  (**Internal Connectors Only**) Add an entry to the `Plugins` map defined in
    `GetInternalPluginsFunc()` of
    [`internal_plugins.go`](../../pkg/secretless/plugin/sharedobj/internal_plugins.go),
    according to your connector type (HTTP/TCP)

1. To test your connector, copy the [`template_connector_test`](template_connector_test)
   directory onto your local machine.

   If you follow the `TODO`-based instructions included in the files in this directory,
   you will be able to write integration tests for your connector using `docker-compose`.
   The included test scripts & files will help you stand up networked containers with
   `docker-compose`.

   **Note for internal connectors:** The the test directory should be copied
   into `test/connector/<connector type>/` and renamed to `<connector_name>`.
   The [`Jenkinsfile`](../../Jenkinsfile) is set up to automatically run the
   integration tests from this directory with each project build.
