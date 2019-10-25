# Using templates to implement Secretless Connector Plugins

We created connector templates to ease the process of adding new connectors to secretless.
Before using the templates to add new connector plugins, be sure to read the [Secretless Connector Plugins README](https://github.com/cyberark/secretless-broker/blob/master/pkg/secretless/plugin/connector/README.md)

To add a new connector do the following:

1. Copy the relevant template directory (HTTP/TCP) into `internal/plugin/connectors/<connector type>`. 
If you're not sure which connector type is suitable, please refer to the [connector technical overview](https://github.com/cyberark/secretless-broker/tree/master/pkg/secretless/plugin/connector#technical-overview)
    1. Inside each template directory you will find the required files & structs implemented, 
    with instructions in the form of TODOs to fill them with the content of the new connector
1.  Add an entry to the `Plugins` map defined in GetInternalPluginsFunc() of
    [`internal_plugins.go`](../../pkg/secretless/plugin/sharedobj/internal_plugins.go), according to their type (HTTP/TCP)
1. Copy the [`template_connector_test`](template_connector_test) directory into `test` and rename it to `<connector_name>_connector`
    1. This directory contains the required test scripts & files, with instructions in the form of TODOs to fill them with your tests
1. Add a new entry to the [`Jenkinsfile`](../../Jenkinsfile) to exercise those test scripts using the `run_integration` script. In most cases, you will also call `junit` on the xml file that `run_integration` outputs in your test's subdirectory.
    Here's an example `Jenkinsfile` entry:
    
    ```
    stage('Integration: PG Handler') {
      steps {
        sh './bin/run_integration pg_handler'
        junit 'test/pg_handler/junit.xml'
      }
    }
    ```
