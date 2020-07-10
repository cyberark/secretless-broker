# Generic HTTP Connector Example Configurations

## Table of Contents
* [Sample Configurations](#sample-configurations)
  * [Docker Registry API](#docker-registry-api)
  * [Elasticsearch API](#elasticsearch-api)
  * [GitHub API](#github-api)
  * [Mailchimp API](#mailchimp-api)
  * [OAuth 2.0 API](#oauth-20-api)
  * [Slack Web API](#slack-web-api)
  * [Splunk API](#splunk-api)
  * [Stripe API](#stripe-api)
  * [Tableau API](#tableau-api)
  * [Twitter API](#twitter-api)
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

> **Protip:** Your target should be either `http://api-target.com` or `api-target.com`.
A URL that starts with https will not work.
___

### Docker Registry API

This example can be used to interact with [Docker's V2 Registry API](https://docs.docker.com/registry/spec/api/#overview).

The configuration file for the Docker Registry API can be found at
[docker_registry_secretless.yml](./docker_registry_secretless.yml).

#### How to use this connector
* Edit the supplied configuration to get your Docker Registry
[Token](https://docs.docker.com/registry/spec/auth/jwt/)
* Run Secretless with the supplied configuration(s)
* Query the Docker Registry API using `http_proxy=localhost:8021 curl <Registry Endpoint URL>/{Request}`

#### Example Usage
<details>
  <summary><b>How to use this connector locally</b></summary>
  <ol>
    <li>Set up a <a href="https://docs.docker.com/registry/deploying/">
    local Registry</a> or use one from <a href="https://hub.docker.com">Dockerhub</a></li>
    <li>Make a request to the Registry API to get the <a href="https://docs.docker.com/registry/spec/auth/oauth/">
    OAuth2 Token</a>.</li>
    <li>
      Store the token from your request in your local credential manager so
      that it may be retrieved in your <code>secretless.yml</code>
    </li>
    <li>Run Secretless locally</li>
    <code>
      ./dist/darwin/amd64/secretless-broker \
      <br />
      -f examples/generic_connector_configs/docker_registry_secretless.yml
    </code>
    <li>List all images from your test Registry using <code>http_proxy=localhost:8021 curl {Registry Endpoint}/v2/repositories/{USERNAME}/?page_size=10000</code></li>
    <li> If you can see the private repos in your repo, you're all set!</li>
  </ol>
</details>

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
  <summary><b>Example setup to try this out locally</b></summary>
  <ol>
    <li>Create an account at <a href="https://cloud.elastic.co/login">
    Elasticsearch's website</a></li>
    <li>Create a <a href="https://www.elastic.co/guide/en/cloud-enterprise/current/ece-restful-api-examples-create-deployment.html">deployment</a></li>
    <li>Make a request to Elasticsearch's API to get the <a href="https://www.elastic.co/guide/en/elasticsearch/reference/master/security-api-get-token.html">
    OAuth2 Token</a>. The request should be made to your deployment Elasticsearch endpoint.</li>
    <li>Run Secretless locally</li>
    <code>
      ./dist/darwin/amd64/secretless-broker \
      <br />
      -f examples/generic_connector_configs/elasticsearch.yml
    </code>
    <li>Query the Elasticsearch API using <code>http_proxy=localhost:9020 curl <Elasticsearch Endpoint URL>/{Request}</code></li>
  </ol>
</details>

___

### GitHub API
This example can be used to interact with [GitHub's API](https://developer.github.com/v3/).

The configuration file for the GitHub API can be found at [github_secretless.yml](./github_secretless.yml).

#### How to use this connector
* Edit the supplied configuration to get your [GitHub OAuth token](https://developer.github.com/v3/#oauth2-token-sent-in-a-header) from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the GitHub API using `http_proxy=localhost:8081 curl api.github.com/{request}`

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally</b></summary>
  <ol>
    <li>
      Get an OAuth token from the Developer Settings page of a user's
      GitHub account
    </li>
    <li>
      Store the token from your request in your local credential manager so
      that it may be retrieved in your <code>secretless.yml</code>
    </li>
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
      On another terminal window, make a request to GitHub using Secretless
    </li>
    <code>
      http_proxy=localhost:8081 curl -X GET api.github.com/users/username
    </code>
  </ol>
</details>

___

### Mailchimp API
This example can be used to interact with [Mailchimp's API](https://mailchimp.com/developer/guides/get-started-with-mailchimp-api-3/).

The configuration file for the Mailchimp API can be found at [mailchimp_secretless.yml](./mailchimp_secretless.yml).

#### How to use this connector
* Edit the supplied configuration to get your Mailchimp OAuth [Access Token](https://mailchimp.com/developer/guides/how-to-use-oauth2/)(OAuth2) or Mailchimp [API Token/Username](https://mailchimp.com/help/about-api-keys/)(Basic Authentication) from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the Mailchimp API using:

```
http_proxy=localhost:{Service IP} curl {dc}.api.mailchimp.com/3.0/{request}
```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally</b></summary>
  <h5>Basic Authentication</h5>
  <ol>
    <li>
      Get an API token from Profile > Extras > API Keys > "Create A Key"
    </li>
    <li>
      Store your username and the token from your request in your local credential manager so
      that it may be retrieved in your <code>mailchimp_secretless.yml</code>
    </li>
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
      On another terminal window, make a request to Mailchimp using Secretless
    </li>
    <code>
      http_proxy=localhost:8010 curl -X GET {dc}.api.mailchimp.com/3.0/
    </code>
  </ol>
  <h5>OAuth2</h5>
  <ol>
    <li>
      Get an Access Token by following the <a href="https://mailchimp.com/developer/guides/how-to-use-oauth2/">provided workflow</a> or by making a Basic Auth API request to <a href="https://mailchimp.com/developer/reference/authorized-apps/">this</a> endpoint 
    </li>
    <li>
      Store the token from your request in your local credential manager so
      that it may be retrieved in your <code>mailchimp_secretless.yml</code>
    </li>
    <li>Build and run Secretless locally</li>
    <code>
      ./bin/build_darwin
    </code>
    <br />
    <code>
    ./dist/darwin/amd64/secretless-broker \
    -f examples/generic_connector_configs/mailchimp_secretless.yml
    </code>
    <li>
        On another terminal window, make a request to Mailchimp using Secretless
    </li>
    <code>
      http_proxy=localhost:8011 curl -X GET {dc}.api.mailchimp.com/3.0/
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
  <summary><b>Example setup to try this out locally...</b></summary>
  <ol>
    <li>Get the Slack <a href="https://slack.com/help/articles/215770388-Create-and-regenerate-API-tokens">application's tokens</a></li>
    <li>
      Store the token from your request in your local credential manager so
      that it may be retrieved in your <code>secretless.yml</code>
    </li>
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
  <summary><b>Example setup to try this out locally</b></summary>
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
    <li>
      Store the token from your request in your local credential manager so
      that it may be retrieved in your <code>secretless.yml</code>
    </li>
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

### Stripe API
This example can be used to interact with
[Stripe's API](https://stripe.com/docs/api).

The configuration file for the Stripe API can be found at
[stripe_secretless.yml](./stripe_secretless.yml).

This example supports several header configurations, so it is recommended to
look at [stripe_secretless.yml](./stripe_secretless.yml) to figure out which
one should be used.

#### How to use this connector
* Get the [Stripe API Key](https://dashboard.stripe.com/apikeys),
which can be used as a Bearer token
* Get a [connected account](https://stripe.com/docs/connect/authentication)
or generate an
[idempotency key](https://stripe.com/docs/api/idempotent_requests) if needed
* Query the Striple API using
`http_proxy=localhost:80*1 curl api.stripe.com/{route}`.

#### Example Usage
<details>
  <summary><b>How to use this connector locally</b></summary>
  <ol>
    <li>Get the Stripe test
      <a href="https://dashboard.stripe.com/apikeys">
        API Key
      </a>
    </li>
    <li>
      Store the token from your request in your local credential manager so
      that it may be retrieved in your <code>secretless.yml</code>
    </li>
    <li>Run Secretless locally</li>
    <code>
    ./dist/darwin/amd64/secretless-broker \
    <br />
    -f examples/generic_connector_configs/stripe_secretless.yml
    </code>
    <li>
      On another terminal window, make a request to Stripe using Secretless
    </li>
    <code>
      http_proxy=localhost:{secretless-server} curl api.stripe.com/v1/charges
    </code>
  </ol>
</details>

___

### Tableau API
This example can be used to interact with
[Tableau's API](https://help.tableau.com/current/api/rest_api/en-us/REST/rest_api.htm).

The configuraton file for the Tableau API can be found at
[tableau_secretless.yml](./tableau_secretless.yml).

#### How to use this connector
* Create an account for Tableau Online
* Make a request to Tableu's API to get an
[`X-Tableau-Auth`](https://help.tableau.com/current/api/rest_api/en-us/REST/rest_api_concepts_auth.htm)
token.
* Run secretless with the supplied configuration(s)
* Query the API using localhost:8071 curl {data} {Tableau Endpoint URl}

#### Example Usage
<details>
  <summary><b>How to use this connector locally</b></summary>
  <ol>
      <li>Create an account on
        <a href="https://www.tableau.com/products/cloud-bi#form">
          Tableau Online
        </a>
      </li>
      <li>
        <a href="https://help.tableau.com/current/api/rest_api/en-us/REST/rest_api_get_started_tutorial_part_1.htm#step-1-sign-in-to-your-server-with-rest">
          Make a POST request
        </a>
        to Tableau Online's API using the provided credentials to secure a
        <code>
          X-Tableau-Auth
        </code>
        token
      </li>
      <li>
        Store the token from your request in your local credential manager so
        that it may be retrieved in your <code>secretless.yml</code>
      </li>
      <li>On another terminal window, make a request to Tableau using Secretless
        <br />
        <code>
          http_proxy=localhost:8071 curl -d {data} {Tableau Endpoint URL}
        </code>
      </li>
    </ol>
</details>

___

### Twitter API
This example can be used to interact with
[Twitter's API](https://developer.twitter.com/en/docs).

The configuration file for the Twitter API can be found at
[twitter_secretless.yml](./twitter_secretless.yml).

**Note:** This configuration currently only supports connecting to the
Twitter API via OAuth2. An issue can be found
[here](https://github.com/cyberark/secretless-broker/issues/1297)
for adding an OAuth1 Connector for Twitter.

#### How to use this connector
* Edit the supplied service configuration to get your
[OAuth token](https://developer.twitter.com/en/docs/basics/authentication/oauth-2-0/bearer-tokens)
* Run secretless with the supplied configuration(s)
* Query the API using `http_proxy=localhost:8051 curl api.twitter.com/{Request}`

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally</b></summary>
  <ol>
    <li>
      Get your
      <a href="https://developer.twitter.com/en/apps">
        Twitter API key and Secret Key
      </a>
    </li>
    <li>
      Get an
      <a href="https://developer.twitter.com/en/docs/basics/authentication/oauth-2-0/bearer-tokens">
        OAuth token
      </a>
      from Twitter through CURL
    </li>
    <code>
      curl -u 'API key:API secret key' \
      <br />
      --data 'grant_type=client_credentials' \
      <br />
      'https://api.twitter.com/oauth2/token'
    </code>
    <li>
      Store the token from your request in your local credential manager so
      that it may be retrieved in your <code>secretless.yml</code>
    </li>
    <li>Run Secretless locally</li>
    <code>
      ./dist/darwin/amd64/secretless-broker \
      <br />
      -f examples/generic_connector_configs/twitter_secretless.yml
    </code>
    <li>
      On another terminal window, make a request to Twitter using Secretless
    </li>
    <code>
      http_proxy=localhost:8051 curl "api.twitter.com/1.1/followers/ids.json?screen_name=twitterdev"
    </code>
  </ol>
</details>

## Contributing

Do you have an HTTP service that you use? Can you write a Secretless generic
connector config for it? **Add the sample config to this folder and list it in
the table above!** Others may find your connector config useful, too - [send us
a PR](https://github.com/cyberark/community/blob/master/CONTRIBUTING.md#contribution-workflow)!
