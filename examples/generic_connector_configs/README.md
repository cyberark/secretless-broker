# Generic HTTP Connector Example Configurations

## Table of Contents
* [Sample Configurations](#sample-configurations)
  * [Elasticsearch API](#elasticsearch-api)
  * [Github API](#github-api)
  * [OAuth 2.0 API](#oauth-20-api)
  * [Slack Web API](#slack-web-api)
  * [Splunk API](#splunk-api)
* [Contributing](#contributing)

## Introduction
The [generic HTTP connector](../../internal/plugin/connectors/http/generic/README.md)
enables using Secretless with a wide array of HTTP-based services _without
having to write new Secretless connectors_. Instead, you can modify your
Secretless configuration to specify the header structure the HTTP service
requires to authenticate.

## Sample Configurations
This section contains a list of generic HTTP configurations that have been built already. Each configuration contains an example of how to run the API locally.

If your target uses self-signed certs you will need to follow the
[documented instructions](https://docs.secretless.io/Latest/en/Content/References/connectors/scl_handlers-https.htm#Manageservercertificates) for adding the
target’s CA to Secretless’ trusted certificate pool.

> Note: The following examples use the [Keychain provider](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/11.3/en/Content/References/providers/scl_keychain.htm?TocPath=Fundamentals%7CSecretless%20Pattern%7CSecret%20Providers%7C_____5).
> Replace the service prefix `service#` with an appropriate service
> or use a different provider as needed.
___
### Elasticsearch API
This example can be used to interact with [Elasticsearch's API](https://www.elastic.co/guide/en/elasticsearch/reference/current).

The configuration file for the Elasticsearch API can be found at
[elasticsearch_secretless.yml](./elasticsearch_secretless.yml).

#### How to use this connector
* Edit the supplied configuration to get your Elasticsearch
[API Key](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html) or
[OAuth token](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-token.html)
* Run Secretless with the supplied configuration(s)
* Query the Elasticsearch API using `http_proxy=localhost:9020 curl <Elasticsearch Endpoint URL>/{Request}`

#### Example Usage
<details>
  <summary><b>How to use this connector locally</b></summary>
  <ol>
    <li>Create an account at <a href="https://cloud.elastic.co/login">
    ElasticSearch's website</a></li>
    <li>Create a <a href="https://www.elastic.co/guide/en/cloud-enterprise/current/ece-restful-api-examples-create-deployment.html">deployment</a></li>
    <li>Make a request to ElasticSearch's API to get the <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/security-api-get-token.html">
    OAuth2 Token</a>. The request should be made to your deployment ElasticSearch endpoint.</li>
    <li>Run Secretless locally</li>
    <code>
      ./dist/darwin/amd64/secretless-broker \
      <br />
      -f examples/generic_connector_configs/elasticsearch.yml
    </code>
    <li>Query the ElasticSearch API using <code>http_proxy=localhost:9020 curl <Elasticsearch Endpoint URL>/{Request}</code></li>
  </ol>
</details>

___

### Github API
This example can be used to interact with [Github's API](https://developer.github.com/v3/).

The configuration file for the Github API can be found at [github_secretless.yml](./github_secretless.yml).

#### How to use this connector
* Edit the supplied configuration to get your [GitHub OAuth token](https://developer.github.com/v3/#oauth2-token-sent-in-a-header) from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the GitHub API using `http_proxy=localhost:8081 curl api.github.com/{request}`

#### Example Usage
<details>
  <summary><b>How to use this connector locally</b></summary>
  <ol>
    <li>
      Get an OAuth token from the Developer Settings page of a user's
      Github account
    </li>
    <li>Added that token into the local machine's OSX Keychain</li>
    <li>Build and run Secretless locally</li>
    <code>
      ./bin/build_darwin
    </code>
    <br />
    <code>
    ./dist/darwin/amd64/secretless-broker \
    -f examples/generic_connector_configs/github_secretless.yml
    </code>
    <li>
      On another terminal window, make a request to Github using Secretless
    </li>
    <code>
      http_proxy=localhost:8081 curl -X GET api.github.com/users/username
    </code>
  </ol>
</details>

___

### OAuth 2.0 API
This generic OAuth HTTP connector can be used for any service that accepts a
Bearer token as an authorization header.

The configuration file for the OAuth 2.0 API can be found at
[oauth2_secretless.yml](./oauth2_secretless.yml).

#### How to use this connector
* Edit the supplied service configuration to get your OAuth token
* Run secretless with the supplied configuration(s)
* Query the API using `http_proxy=localhost:8071 curl <Your OAuth2 API Endpoint URL>/{Request}`

___

### Slack Web API
This example can be used to interact with [Slack's Web API](https://api.slack.com/apis).

The configuration file for the Slack Web API can be found at [slack_secretless.yml](./slack_secretless.yml).

#### How to use this connector
* Edit the supplied configuration to get your Slack [OAuth token](https://api.slack.com/legacy/oauth#flow)
* Run secretless with the supplied configuration(s)
* Query the Slack API using `http_proxy=localhost:9030 curl -d {data} <Slack Endpoint URL>` or `http_proxy=localhost:9040 curl -d {data} <Slack Endpoint URL>`
depending on if your endpoint requires JSON or URL encoded requests

#### Example Usage
<details>
  <summary><b>How to use this connector locally...</b></summary>
  <ol>
    <li>Get the Slack <a href="https://slack.com/help/articles/215770388-Create-and-regenerate-API-tokens">application's tokens</a></li>
    <li>Save the local token from Slack into the OSX keychain</li>
    <li>Run Secretless locally</li>
    <code>
      ./dist/darwin/amd64/secretless-broker \
      <br />
      -f examples/generic_connector_configs/slack_secretless.yml
    </code>
    <li>On another terminal window, make a request to Slack using Secretless</li>
    <code>http_proxy=localhost:9030 curl -X POST --data '{"channel":"C061EG9SL","text":"I hope the tour went well"}' slack.com/api/chat.postMessage</code>
  </ol>
</details>

___

### Splunk API
This example can be used to interact with [Splunk's API](https://api.slack.com/apis).

The configuration file for the Splunk Web API can be found at [splunk_secretless.yml](./splunk_secretless.yml).

#### How to use this connector
* Edit the supplied configuration to get your [Splunk authentication token](https://docs.splunk.com/Documentation/Splunk/8.0.2/Security/EnableTokenAuth)
from the correct provider/path
* Create a Splunk [certficate](https://docs.splunk.com/Documentation/Splunk/8.0.2/Security/Howtoself-signcertificates) and add the certificate to [Secretless's trusted certificate pool](https://docs.secretless.io/Latest/en/Content/References/connectors/scl_handlers-https.htm#Manageservercertificates)
* Run Secretless with the supplied configuration
* Query the Splunk API using `http_proxy=localhost:8081 curl {instance host name or IP address}:{management port}/{route}` - note that you do not preface your
instance host name with `https://`; Secretless will ensure the final connection
to the backend server uses SSL.

#### Example Usage
<details>
  <summary><b>How to use this connector locally</b></summary>
  <ol>
    <li>Run a local instance of Splunk in a Docker container</li>
    <code>
    docker run \
      <br />
        -d \
      <br />
        -p 8000:8000 \
      <br />
        -p 8089:8089 \
      <br />
        -e "SPLUNK_START_ARGS=--accept-license" \
      <br />
        -e "SPLUNK_PASSWORD=specialpass" \
      <br />
        --name splunk \
      <br />
        splunk/splunk:latest
    </code>
    <li>
      Follow the instructions
      <a href="https://docs.splunk.com/Documentation/Splunk/8.0.2/Security/EnableTokenAuth">here</a>
      to create a local Splunk token using Splunk Web
    </li>
    <li>Save the local token from Splunk Web into the OSX keychain</li>
    <li>
      Add 'SplunkServerDefaultCert' at IP 127.0.0.1 to etc/hosts on the machine.
      This was so the host name of the HTTP Request would match the name on the
      certificate that is provided on our Splunk container
    </li>
    <li>
      Use the provided cacert.pem file on the Splunk docker container
      for my certificate, and write it to the local machine
    </li>
    <code>docker exec -it splunk sudo cat /opt/splunk/etc/auth/cacert.pem > myLocalSplunkCertificate.pem</code>
    <li>
      Set a variable in the terminal named
      <code>
        <a href="https://docs.conjur.org/latest/en/Content/References/connectors/scl_handlers-https.htm?TocPath=Fundamentals%7CSecretless%20Pattern%7CService%20Connectors%7CHTTP%7C_____0">SECRETLESS_HTTP_CA_BUNDLE</a>
      </code>
      and set it to the path where myLocalSpunkCertificate.pem was
      on the local machine.
    </li>
    <li>Run Secretless</li>
    <code>
      ./dist/darwin/amd64/secretless-broker \
      <br />
      -f examples/generic_connector_configs/splunk_secretless.yml
    </code>
    <li>On another terminal window, make a request to Splunk using Secretless</li>
    <code>http_proxy=localhost:8081 curl -k -X GET SplunkServerDefaultCert:8089/services/apps/local</code>
  </ol>
</details>

___

## Contributing

Do you have an HTTP service that you use? Can you write a Secretless generic
connector config for it? **Add the sample config to this folder and list it in
the table above!** Others may find your connector config useful, too - [send us
a PR](https://github.com/cyberark/community/blob/master/CONTRIBUTING.md#contribution-workflow)!
