# Using templates to implement Secretless Connector Plugins

We created connector templates to ease the process of adding new connectors to secretless.
Before using the templates to add new connector plugins, be sure to read the [Secretless Connector Plugins README](https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md)

To add a new connector do the following:

1. Copy the relevant template directory (HTTP/TCP) into `internal/plugin/connectors/<connector type>`. 
If you're not sure which connector type is suitable, please refer to the [connector technical overview](https://github.com/cyberark/secretless-broker/tree/master/pkg/secretless/plugin/connector#technical-overview).
    1. Inside each template directory you will find the required files & structs implemented, 
    with instructions in the form of TODOs to fill them with the content of the new connector.
1.  Add an entry to the `Plugins` map defined in GetInternalPluginsFunc() of
    [`internal_plugins.go`](../../pkg/secretless/plugin/sharedobj/internal_plugins.go), according to their type (HTTP/TCP)
1. Copy the [`template_connector_test`](template_connector_test) directory into `test/connector/<connector type>/` and rename it to `<connector_name>`.
    1. This directory will help you write integration tests for your connector. It contains test scripts & files to help you stand up networked containers with docker-compose. The files give instructions on the steps to set up your test suite in the form of TODOs.
       The [`Jenkinsfile`](../../Jenkinsfile) is set up to automatically run the integration tests with each project build.
