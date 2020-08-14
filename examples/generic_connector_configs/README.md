# Generic HTTP Connector Example Configurations

## Table of Contents
* [Sample Configurations](#sample-configurations)
  * [Datadog API](#datadog-api)
  * [Docker Registry API](#docker-registry-api)
  * [Dropbox API](#dropbox-api)
  * [Elasticsearch API](#elasticsearch-api)
  * [Facebook API](#facebook-api)
  * [Full Contact API](#full-contact-api)
  * [GitHub API](#github-api)
  * [Google Maps API](#google-maps-api)
  * [JFrog Artifactory API](#jfrog-artifactory-api)
  * [Logentries API](#logentries-api)
  * [Loggly API](#loggly-api)
  * [Mailchimp API](#mailchimp-api)
  * [New Relic API](#new-relic-api)
  * [OAuth 1.0 API](#oauth-10-api)
  * [OAuth 2.0 API](#oauth-20-api)
  * [Papertrail API](#papertrail-api)
  * [SendGrid Web API](#sendgrid-web-api)
  * [Sentry API](#sentry-api)
  * [Service Now API](#service-now-api)
  * [Slack Web API](#slack-web-api)
  * [Splunk API](#splunk-api)
  * [Stripe API](#stripe-api)
  * [Tableau API](#tableau-api)
  * [Twilio API](#twilio-api)
  * [Twitter API](#twitter-api)
* [Contributing](#contributing)

## Introduction
The
[generic HTTP connector](../../internal/plugin/connectors/http/generic/README.md)
enables using Secretless with a wide array of HTTP-based services _without
having to write new Secretless connectors_. Instead, you can modify your
Secretless configuration to specify the header structure the HTTP service
requires to authenticate.

## Sample Configurations
This section contains a list of generic HTTP configurations that have been
built already. Each configuration contains an example of how to run the API
locally.

If your target uses self-signed certs you will need to follow the
[documented instructions](https://docs.secretless.io/Latest/en/Content/References/connectors/scl_handlers-https.htm#Manageservercertificates) for adding the
target’s CA to Secretless’ trusted certificate pool.

> Note: The following examples use the [Keychain provider](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/11.3/en/Content/References/providers/scl_keychain.htm?TocPath=Fundamentals%7CSecretless%20Pattern%7CSecret%20Providers%7C_____5).
> Replace the service prefix `service#` with an appropriate service
> or use a different provider as needed.

> **Protip:** Your target should be either `http://api-target.com` or
`api-target.com`. A URL that starts with https will not work.
___

### Datadog API
This example can be used to interact with
[Datadog API](https://docs.datadoghq.com/api/v2/).

The configuration file for the Datadog API can be found at
[datadog_secretless.yml](./datadog_secretless.yml).

> This configuration uses [v2](https://docs.datadoghq.com/api/v2/)
of the DataDog API.

#### How to use this connector

* Edit the supplied configuration to get your Datadog
[API Key](https://docs.datadoghq.com/account_management/api-app-keys/)
or/and
[Application Key](https://docs.datadoghq.com/account_management/api-app-keys/)
* Run Secretless with the supplied configuration(s)
* Query the Datadog API using

  ```
  http_proxy=localhost:8041 curl api.datadoghq.com/{request}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Set up a [Datadog Account](https://app.datadoghq.com/) and get an
     [API Key](https://docs.datadoghq.com/account_management/api-app-keys/)
  1. Get a DataDog
     [Application key](https://docs.datadoghq.com/account_management/api-app-keys/)
  1. Store the token from your request in your local credential manager so
  that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
     ./dist/darwin/amd64/secretless-broker \
     -f examples/generic_connector_configs/datadog_secretless.yml
     ```
  1. Query the API using
  `http_proxy=localhost:8041 curl api.datadoghq.com/api/v1/user`

</details>

___

### Docker Registry API

This example can be used to interact with
[Docker's V2 Registry API](https://docs.docker.com/registry/spec/api/#overview).

The configuration file for the Docker Registry API can be found at
[docker_registry_secretless.yml](./docker_registry_secretless.yml).

> This configuration uses v2 of the Docker Registry API.

#### How to use this connector

* Edit the supplied configuration to get your Docker Registry
[Token](https://docs.docker.com/registry/spec/auth/jwt/)
* Run Secretless with the supplied configuration(s)
* Query the Docker Registry API using

  ```
  http_proxy=localhost:8021 curl <Registry Endpoint URL>/{Request}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Set up a [local Registry](https://docs.docker.com/registry/deploying/)
  or use one from [Dockerhub](https://hub.docker.com)
  1. Make a request to the Registry API to get the
  [OAuth2 Token](https://docs.docker.com/registry/spec/auth/oauth/).
  1. Store the token from your request in your local credential manager so
  that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
     ./dist/darwin/amd64/secretless-broker \
     -f examples/generic_connector_configs/docker_registry_secretless.yml
     ```
  1. List all images from your test Registry using
  `http_proxy=localhost:8021 curl {Registry Endpoint}/v2/repositories/{USERNAME}/?page_size=10000`
  1. If you can see the private repos in your repo, you're all set!

</details>

___

### Dropbox API
This example can be used to interact with
[Dropbox's API](https://www.dropbox.com/developers/documentation/http/overview).

The configuration file for the Dropbox API can be found at
[dropbox_secretless.yml](./dropbox_secretless.yml).

> This configuration uses v2 of the Dropbox API.

#### How to use this connector

* Edit the supplied configuration to get your
[Dropbox API token](https://www.dropbox.com/developers/apps) or
[App key and App Secret](https://www.dropbox.com/developers/apps)
* Run Secretless with the supplied configuration
* Query the Dropbox API using
`http_proxy=localhost:8081 curl {Dropbox Route}/{Request}`

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally</b></summary>

  1. Get an OAuth token from the settings page of a
     [Dropbox application](https://www.dropbox.com/developers/apps)
  1. Store the token from your request in your local credential manager so
      that it may be retrieved in your `secretless.yml`
  1. Build and run Secretless locally
     ```
     ./bin/build_darwin
     ./dist/darwin/amd64/secretless-broker \
          -f examples/generic_connector_configs/dropbox_secretless.yml
     ```
  1. On another terminal window, make a request to Dropbox using Secretless
     ```
     http_proxy=localhost:8081 curl -X POST api.dropboxapi.com/2/team/get_info
     ```
</details>

___

### Elasticsearch API
This example can be used to interact with
[Elasticsearch's API](https://www.elastic.co/guide/en/elasticsearch/reference/current).

The configuration file for the Elasticsearch API can be found at
[elasticsearch_secretless.yml](./elasticsearch_secretless.yml).

> This configuration uses v7.8 of the Elasticsearch API.

#### How to use this connector

* Edit the supplied configuration to get your Elasticsearch
[API Key](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html)
or
[OAuth token](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-token.html)
* Run Secretless with the supplied configuration(s)
* Query the Elasticsearch API using

  ```
  http_proxy=localhost:9020 curl {Elasticsearch Endpoint URL}/{Request}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Create an account at
  [Elasticsearch's website](https://cloud.elastic.co/login)
  1. Create a
  [deployment](https://www.elastic.co/guide/en/cloud-enterprise/current/ece-restful-api-examples-create-deployment.html)
  1. Make a request to Elasticsearch's API to get the
  [OAuth2 Token](https://www.elastic.co/guide/en/elasticsearch/reference/master/security-api-get-token.html).
  The request should be made to your deployment Elasticsearch endpoint.
  1. Run Secretless locally
     ```
     ./dist/darwin/amd64/secretless-broker \
     -f examples/generic_connector_configs/elasticsearch.yml
     ```
  1. Query the Elasticsearch API using
     ```
     http_proxy=localhost:9020 curl {Elasticsearch Endpoint URL}/{Request}
     ```

</details>

___

### Facebook API

This example focuses on [Instagram](https://developers.facebook.com/docs/instagram),
but can be used to interact with
[Facebook's many API's](https://developers.facebook.com/docs).

The configuration file for the Instagram API can be found at
[instagram_secretless.yml](./instagram_secretless.yml).

> This configuration was created using v2020-05-05. You can read about
what changes have been made in the
[Instagram API Changelog](https://developers.facebook.com/docs/instagram-api/changelog)

#### How to use this connector

* Edit the supplied configuration to get your
  [OAuth token](https://developers.facebook.com/docs/pages/access-tokens)
  from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the GitHub API using `http_proxy=localhost:8023 curl {host}/{request}`

#### Example Usage

<details>
  <summary><b>Example setup to try this out locally</b></summary>
  This example focuses on Instagram, but using Secretless with the rest of the
  Facebook API's is the same from Step 2 on.

  1. Follow this guide to set up an
   [example Instagram app](https://developers.facebook.com/docs/instagram-basic-display-api/getting-started)
   and acquire an access token
  1. Store the token in your local credential manager so
   that it may be retrieved in your `secretless.yml`
  1. Build and run Secretless locally
     ```
       ./bin/build_darwin
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/instagram_secretless.yml
     ```
  1. In another terminal window, make a request to Instagram using Secretless
     ```
        http_proxy=localhost:8023 curl -X GET --url 'graph.instagram.com/me?fields=id,username'
     ```
</details>

___

### Full Contact API

This example can be used to interact with
[Full Contact's API](https://dashboard.fullcontact.com/api-ref).

The configuration file for the Full Contact API can be used at
[full_contact_secretless.yml](./full_contact_secretless.yml).

> This configuration can be used for both v2 and v3 of the Full Contact API.

#### How to use this connector

* Edit the supplied configuration to get your
[Full Contact API Key/Bearer Token](https://dashboard.fullcontact.com/)
from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the Full Contact API using:

  **X-FullContact-APIKey:**
  ```
    http_proxy=localhost:8081 curl
      'api.fullcontact.com/v2/person.json?email=bart@fullcontact.com'
  ```

  **OAuth 2.0:**
  ```
  http_proxy=localhost:8071 curl -X POST api.fullcontact.com/v3/person.enrich \
    -H "Content-Type: application/json" \
    -d '{"email":"bart@fullcontact.com"}'
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get an API key from "Get an API Key" on the Full Contact dashboard.
     Regardless of which version of the Full Contact API you are using, you can
     use an API key as both a token and API key.
  1. Store your API Key from your request in your local
     credential manager so that it may be retrieved in your
     `full_contact_secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/full_contact_secretless.yml
     ```
  1. On another terminal window, make a request to Mailchimp using Secretless
     ```
      http_proxy=localhost:8071 curl -X POST \
        api.fullcontact.com/v3/person.enrich \
        -H "Content-Type: application/json"  \
        -d '{"email":"bart@fullcontact.com"}'
     ```

</details>

___

### GitHub API

This example can be used to interact with
[GitHub's API](https://developer.github.com/).

The configuration file for the GitHub API can be found at
[github_secretless.yml](./github_secretless.yml).

> This configuration uses v3 of the Github API.

#### How to use this connector
* Edit the supplied configuration to get your
[GitHub OAuth token](https://developer.github.com/v3/#oauth2-token-sent-in-a-header)
from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the GitHub API using

  ```
  http_proxy=localhost:8081 curl api.github.com/{request}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get an OAuth token from the Developer Settings page of a user's
  GitHub account
  1. Store the token from your request in your local credential manager so
  that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
        ./dist/darwin/amd64/secretless-broker \
        -f examples/generic_connector_configs/github_secretless.yml
     ```
  1. On another terminal window, make a request to GitHub using Secretless
     ```
        http_proxy=localhost:8081 curl -X GET api.github.com/users/username
     ```

</details>

___

### Google Maps API

This example can be used to interact with the
[Google Maps Web Services APIs](https://developers.google.com/maps/apis-by-platform).

The configuration file for the GitHub API can be found at
[google_maps_secretless.yml](./google_maps_secretless.yml).

> This configuration was created on July 21, 2020. You can read about
what changes have been made in the
[Google API changelog](https://cloud.google.com/maps-platform/user-guide/product-changes#maps)

#### How to use this connector
* Edit the supplied configuration to get your
[Google API Key](https://developers.google.com/maps/documentation/javascript/get-api-key)
from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the Google Maps Web Services API using

  ```
  http_proxy=localhost:8081 curl {Google Maps Route}/{request}/{params}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get a Google API key from the
     [Google Console](https://console.developers.google.com/apis/credentials)
  1. Store the token from your request in your local credential manager so
  that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
        ./dist/darwin/amd64/secretless-broker \
        -f examples/generic_connector_configs/google_maps_secretless.yml
     ```
  1. On another terminal window, make a request to the Google Maps Web Service
     API using Secretless
     ```
        http_proxy=localhost:8081 curl -X POST www.googleapis.com/geolocation/v1/geolocate
     ```

</details>

___

### JFrog Artifactory API

This example can be used to interact with the
[JFrog Artifactory API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API).

The configuration file for the GitHub API can be found at
[jfrog_artifactory_secretless.yml](./jfrog_artifactory_secretless.yml).

> This configuration uses
[v6.x](https://www.jfrog.com/confluence/display/RTF6X/Artifactory+REST+API)
of the JFrog Artifactory REST Api.

#### How to use this connector
* Edit the supplied configuration to get your
[JFrog Artifactory API Key](https://www.jfrog.com/confluence/display/JFROG/User+Profile#UserProfile-APIKey),
[JFrog Artifactory Access Token](https://www.jfrog.com/confluence/display/JFROG/Access+Tokens)
or JFrog Artifactory Username/Password.
* Run Secretless with the supplied configuration
* Query the JFrog Artifactory API using:

  ```
  http_proxy=localhost:8071 curl {JFrog Host Name}.jfrog.io/router/api/v1/system/ping
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  **Basic Authentication**
  1. Get a JFrog Artifactory Account
  1. Store your username and the password/API key/token from your request in
     your local credential manager so that it may be retrieved in your
     `jfrog_artifactory_secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/jfrog_artifactory_secretless.yml
     ```
  1. On another terminal window, make a request to JFrog using Secretless
     ```
       http_proxy=localhost:8071 curl {JFrog Host Name}.jfrog.io/router/api/v1/system/ping
     ```

  **API Key Authentication**
  1. Get a JFrog Artifactory
     [API key](https://www.jfrog.com/confluence/display/JFROG/User+Profile#UserProfile-APIKey)
  1. Store your API key from your request in your local
     credential manager so that it may be retrieved in your
     `jfrog_artifactory_secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/jfrog_artifactory_secretless.yml
     ```
  1. On another terminal window, make a request to JFrog using Secretless
     ```
       http_proxy=localhost:8081 curl {JFrog Host Name}.jfrog.io/router/api/v1/system/ping
     ```

  **Auth Token Authentication**
  1. Get a JFrog Artifactory
     [access token](https://www.jfrog.com/confluence/display/JFROG/Access+Tokens)
  1. Store your username and the token from your request in your local
     credential manager so that it may be retrieved in your
     `jfrog_artifactory_secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/jfrog_artifactory_secretless.yml
     ```
  1. On another terminal window, make a request to JFrog using Secretless
     ```
       http_proxy=localhost:8091 curl {JFrog Host Name}.jfrog.io/router/api/v1/system/ping
     ```

</details>

___

### Logentries API

This example can be used to interact with the
[Logentries InsightOps API](https://docs.rapid7.com/insightops/rest-api-overview).

The configuration file for the Logentries API can be found at
[logentries_secretless.yml](./logentries_secretless.yml).

> This configuration was made on July 28th, 2020. Logentries does not specify
their API version, so the example usage may change in the future.

#### How to use this connector
* Edit the supplied configuration to get your
[Logentries API Key](https://docs.rapid7.com/insightops/rest-api-overview#obtain-an-api-key)
* Run Secretless with the supplied configuration
* Query the Logentries API using:

  ```
  http_proxy=localhost:8071 curl {Country Name}.rest.logs.insight.rapid7.com/{Route}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Login to create a [Logentries API Key](https://insight.rapid7.com/platform#/apiKeyManagement)
  1. Store the token from your request in your local credential manager so
  that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
        ./dist/darwin/amd64/secretless-broker \
        -f examples/generic_connector_configs/logentries_secretless.yml
     ```
  1. On another terminal window, make a request to the Logentries API using
     Secretless
     ```
        http_proxy=localhost:8071 curl us.rest.logs.insight.rapid7.com/query/saved_queries
     ```

</details>

___

### Loggly API

This example can be used to interact with the
[Loggly API](https://documentation.solarwinds.com/en/Success_Center/loggly/Content/admin/api-overview.htm).

The configuration file for the Loggly API can be found at
[loggly_secretless.yml](./loggly_secretless.yml).

> This configuration was created on July 27, 2020. The release notes for the
Loggly API can be found [here](https://documentation.solarwinds.com/en/Success_Center/loggly/Content/Release_Notes/release_notes.htm).

> Note: This configuration does not support sending events to the Loggly API.
To see how to send events to the Loggly API, view their
[documentation](https://documentation.solarwinds.com/en/Success_Center/loggly/Content/admin/api-sending-data.htm).

#### How to use this connector

* Edit the supplied configuration to get your
  [Loggly token](https://documentation.solarwinds.com/en/Success_Center/loggly/Content/admin/token-based-api-authentication.htm)
  or your Loggly username and password
* Run Secretless with the supplied configuration
* Query the Loggly API using:
  ```
  http_proxy=localhost:8071 curl -v '{your-loggly-link}.loggly.com/apiv2/events/iterate?q=*&from=-10m&until=now&size=10'
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get a
     [Loggly token](https://documentation.solarwinds.com/en/Success_Center/loggly/Content/admin/token-based-api-authentication.htm)
     or authenticate using basic authorization.
  1. Store your username and the token from your request in your local
     credential manager so that it may be retrieved in your
     `loggly_secretless.yml`
  1. Run Secretless locally
     ```
      ./dist/darwin/amd64/secretless-broker \
      -f examples/generic_connector_configs/loggly_secretless.yml
     ```
  1. Query the Loggly API using:
     ```
      http_proxy=localhost:8071 curl '{your-loggly-link}.loggly.com/apiv2/events/iterate?q=*&from=-10m&until=now&size=10'
     ```

</details>

___

### Mailchimp API

This example can be used to interact with
[Mailchimp's API](https://mailchimp.com/developer/guides/get-started-with-mailchimp-api-3/).

The configuration file for the Mailchimp API can be found at
[mailchimp_secretless.yml](./mailchimp_secretless.yml).

> This configuration uses v3 of the Mailchimp API.

#### How to use this connector

* Edit the supplied configuration to get your Mailchimp OAuth
  [Access Token](https://mailchimp.com/developer/guides/how-to-use-oauth2/)
  OAuth2 or Mailchimp
  [API Token/Username](https://mailchimp.com/help/about-api-keys/)
  Basic Authentication from the correct provider/path.
* Run Secretless with the supplied configuration
* Query the Mailchimp API using:

  ```
  http_proxy=localhost:{Service IP} curl {dc}.api.mailchimp.com/3.0/{request}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  **Basic Authentication**

  1. Get an API token from Profile > Extras > API Keys > "Create A Key"
  1. Store your username and the token from your request in your local
     credential manager so that it may be retrieved in your
     `mailchimp_secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/mailchimp_secretless.yml
     ```
  1. On another terminal window, make a request to Mailchimp using Secretless
     ```
       http_proxy=localhost:8010 curl -X GET {dc}.api.mailchimp.com/3.0/
     ```

  **OAuth2**

  1. Get an Access Token by following the
  [provided workflow](https://mailchimp.com/developer/guides/how-to-use-oauth2/)
  or by making a Basic Auth API request to
  [this](https://mailchimp.com/developer/reference/authorized-apps/) endpoint
  1. Store your username and the token from your request in your local
  credential manager so that it may be retrieved in your
  `mailchimp_secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/mailchimp_secretless.yml
     ```
  1. On another terminal window, make a request to Mailchimp using Secretless
    ```
      http_proxy=localhost:8011 curl -X GET {dc}.api.mailchimp.com/3.0/
    ```

</details>

___

### New Relic API

This example can be used to interact with
[New Relic's API](https://docs.newrelic.com/docs/apis/rest-api-v2).

The configuration file for the New Relic API can be found at
[new_relic_secretless.yml](./new_relic_secretless.yml).

> This configuration uses [v2](https://docs.newrelic.com/docs/apis/rest-api-v2)
> of the New Relic API.

#### How to use this connector

* Edit the supplied configuration to get your [New Relic API Key](https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys)
* Run Secretless with the supplied configuration
* Query the New Relic API using:

  ```
  http_proxy=localhost:8071 curl api.newrelic.com/v2/applications.json
  ```

#### Example Usage

<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get a
     [New Relic API Key](https://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys)
  1. Store the token from your request in your local credential manager so
     that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
      ./dist/darwin/amd64/secretless-broker \
      -f examples/generic_connector_configs/new_relic_secretless.yml
     ```
  1. Query the New Relic API using:
     ```
      http_proxy=localhost:8071 curl api.newrelic.com/v2/{Route}
     ```

</details>

___

### OAuth 1.0 API

This generic OAuth HTTP connector can be used for any service that uses OAuth1
for authorization.

The configuration file for connecting to an API that uses OAuth 1.0 can be found
at [oauth1_secretless.yml](./oauth1_secretless.yml).

> Note: Secretless currently only supports HMAC-SHA1 hashing. There is an
> [issue](https://github.com/cyberark/secretless-broker/issues/1324)
> logged to support other hashing algorithms such as RSA-SHA1 and PLAINTEXT.

#### How to use this connector
* Edit the supplied service configuration to get your OAuth token

* Run Secretless with the supplied configuration(s)

* Query the API using
  ```
  http_proxy=localhost:8071 curl {Your OAuth1 API Endpoint URL}/{Request}
  ```

___

### OAuth 2.0 API

This generic OAuth HTTP connector can be used for any service that accepts a
Bearer token as an authorization header.

The configuration file for connecting to an API that uses OAuth 2.0 can be found
at [oauth2_secretless.yml](./oauth2_secretless.yml).

#### How to use this connector
* Edit the supplied service configuration to get your OAuth token

* Run Secretless with the supplied configuration(s)

* Query the API using
  ```
  http_proxy=localhost:8071 curl {Your OAuth2 API Endpoint URL}/{Request}
  ```

___

### Papertrail API

This example can be used to interact with the
[Papertrail API](https://help.papertrailapp.com/kb/how-it-works/http-api/).

The configuration file for the Papertrail API can be found at
[papertrail_secretless.yml](./papertrail_secretless.yml).

#### How to use this connector

* Edit the supplied configuration to get your
[API Key](https://papertrailapp.com/account/profile)

* Run Secretless with the supplied configuration(s)

* Query the API using
  ```
  http_proxy=localhost:8071 curl papertrailapp.com/api/v1/{Request}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Retrieve your API token from your Profile
  1. Store the token from your request in your local credential manager so
     that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
     ./dist/darwin/amd64/secretless-broker \
     -f examples/generic_connector_configs/papertrail_secretless.yml
     ```
  1. In another terminal window, make a request to Papertrail using Secretless
     ```
      http_proxy=localhost:8043 curl papertrailapp.com/api/v1/systems.json
     ```

</details>

___

### SendGrid Web API

This example can be used to interact with the
[SendGrid Web API](https://sendgrid.com/docs/API_Reference/api_v3.html).

The configuration file for the SendGrid Web API can be found at
[sendgrid_secretless.yml](./sendgrid_secretless.yml).

> This configuration uses [v3](https://sendgrid.com/docs/API_Reference/api_v3.html)
> of the SendGrid Web API.

#### How to use this connector

* Edit the supplied configuration to get your
[SendGrid API Key](https://app.sendgrid.com/settings/api_keys)

* Run Secretless with the supplied configuration(s)

* Query the API using
  ```
  http_proxy=localhost:8071 curl api.sendgrid.com/{Request}
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Generate a
     [SendGrid API Key](https://sendgrid.api-docs.io/v3.0/how-to-use-the-sendgrid-v3-api/api-authentication)
  1. Store the token from your request in your local credential manager so
    that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
     ./dist/darwin/amd64/secretless-broker \
     -f examples/generic_connector_configs/sendgrid_secretless.yml
     ```
  1. On another terminal window, make a request to SendGrid using Secretless
     ```
       http_proxy=localhost:8071 curl --request POST \
       --url api.sendgrid.com/v3/mail/send \
       --header 'Content-Type: application/json' \
       --data '{"personalizations": [{"to": [{"email": "test@example.com"}]}],
       "from": {"email": "test@example.com"},"subject": "Sending with SendGrid
       is Fun","content": [{"type": "text/plain", "value": "and easy to do
       anywhere, even with cURL"}]}'
     ```

</details>

___

### Sentry API

This example can be used to interact with
the [Sentry API](https://docs.sentry.io/api/).

The configuration file for the Sentry API can be found at
[sentry_secretless.yml](./sentry_secretless.yml).

> This configuration uses v0 of the Sentry API.

#### How to use this connector

* Edit the supplied configuration to get your Sentry
  [token](https://sentry.io/settings/account/api/auth-tokens/)

* Run Secretless with the supplied configurations

* Query the Sentry API
  ```
  http_proxy=localhost:8071 curl sentry.io/api/0/projects/
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get a Sentry
  [token](https://sentry.io/settings/account/api/auth-tokens/)
  1. Store the token from your request in your local credential manager so that
  it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
     ./dist/darwin/amd64/secretless-broker \
     -f examples/generic_connector_configs/sentry_secretless.yml
     ```
  1. On another terminal window, make a request to Sentry using Secretless
     ```
     http_proxy=localhost:8071 curl sentry.io/api/0/projects/
     ```

</details>

___

### Service Now API

This example can be used to interact with
[Service Now's API](https://docs.servicenow.com/bundle/orlando-application-development/page/build/applications/concept/api-rest.html)

The configuration file for the Service Now API can be found at
[service_now_secretless.yml](./service_now_secretless.yml).

> This configuration uses v2 of the Service Now API.

#### How to use this connector

* Edit the supplied configruation to get your Service Now OAuth
  [Token](https://hi.service-now.com/kb_view.do?sysparm_article=KB0725643)

* Run Secretless with the supplied configuration(s)

* Query the Service Now API
  ```
    http_proxy=localhost:8071 curl  {instance-name}.service-now.com/api/now/v2/table/incident
  ```

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Go to your Authorization Endpoint from the
     [authentication documentation](https://hi.service-now.com/kb_view.do?sysparm_article=KB0725643)
     ```
      https://{instance}.service-now.com/oauth_auth.do?response_type=code&redirect_uri=https://{instance}.service-now.com/login.do&client_id={client-id}
     ```
  1. Grab the `code` parameter
  1. Make a curl request to get your authorization endpoint.
     ```
     curl --location --request POST \
      'https://{instance-name}.service-now.com/oauth_token.do' \
      --header 'Content-Type: application/x-www-form-urlencoded' \
      --data-urlencode 'grant_type=password' \
      --data-urlencode 'code={code}' \
      --data-urlencode 'client_id={client-id}' \
      --data-urlencode 'client_secret={client-secret}}' \
      --data-urlencode 'redirect_uri={redirect_uri}' \
      --data-urlencode 'username={username}' \
      --data-urlencode 'password={password}'
     ```
   1. Store the token from your request in your local credential manager so that
      it may be retrieved in your `secretless.yml`
   1. Run Secretless locally
      ```
        ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/service_now_secretless.yml
      ```
   1. On another terminal window, make a request to Slack using Secretless
      ```
        http_proxy=localhost:8071 curl  {instance}.service-now.com/api/now/v2/table/incident
      ```

</details>

___

### Slack Web API

This example can be used to interact with
[Slack's Web API](https://api.slack.com/apis).

The configuration file for the Slack Web API can be found at
[slack_secretless.yml](./slack_secretless.yml).

> This configuration was created on June 22, 2020. You can read about
what changes have been made in the
[Slack changelog](https://api.slack.com/changelog)

#### How to use this connector

* Edit the supplied configuration to get your Slack
  [OAuth token](https://api.slack.com/legacy/oauth#flow)

* Run Secretless with the supplied configuration(s)

* Query the Slack API
  ```
  http_proxy=localhost:9030 curl -d {data} {Slack Endpoint URL}
  ```
  ```
  http_proxy=localhost:9040 curl -d {data} {Slack Endpoint URL}
  ```
  Your query depends on if your endpoint requires JSON or URL encoded requests.

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get the Slack
  [application's tokens](https://slack.com/help/articles/215770388-Create-and-regenerate-API-tokens)
  1. Store the token from your request in your local credential manager so that
  it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/slack_secretless.yml
     ```
  1. On another terminal window, make a request to Slack using Secretless
     ```
       http_proxy=localhost:9030 curl -X POST --data '{"channel":"C061EG9SL",
       "text":"I hope the tour went well"}' slack.com/api/chat.postMessage
     ```

</details>

___

### Splunk API

This example can be used to interact with
[Splunk's API](https://docs.splunk.com/Documentation/Splunk/8.0.5/RESTREF/RESTprolog).

The configuration file for the Splunk Web API can be found at
[splunk_secretless.yml](./splunk_secretless.yml).

> This configuration uses v8.0.5 of the SendGrid Web API.

#### How to use this connector

* Edit the supplied configuration to get your
  [Splunk authentication token](https://docs.splunk.com/Documentation/Splunk/8.0.2/Security/EnableTokenAuth)
  from the correct provider/path
* Create a Splunk
  [certficate](https://docs.splunk.com/Documentation/Splunk/8.0.2/Security/Howtoself-signcertificates)
  and add the certificate to
  [Secretless's trusted certificate pool](https://docs.secretless.io/Latest/en/Content/References/connectors/scl_handlers-https.htm#Manageservercertificates)
* Run Secretless with the supplied configuration
* Query the Splunk API using

  ```
  http_proxy=localhost:8081 curl {instance host name or IP address}:{management port}/{route}
  ```
> Note: You do not preface your instance host name with `https://`.
Secretless will ensure the final connection to the backend server uses SSL.

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Run a local instance of Splunk in a Docker container
     ```
     docker run \
         -d \
         -p 8000:8000 \
         -p 8089:8089 \
         -e "SPLUNK_START_ARGS=--accept-license" \
         -e "SPLUNK_PASSWORD=specialpass" \
         --name splunk \
         splunk/splunk:latest
     ```
  1. Follow the instructions
  [here](https://docs.splunk.com/Documentation/Splunk/8.0.2/Security/EnableTokenAuth)
  to create a local Splunk token using Splunk Web
  1. Store the token from your request in your local credential manager so that
  it may be retrieved in your `secretless.yml`</code>`
  1. Add 'SplunkServerDefaultCert' at IP 127.0.0.1 to etc/hosts on the machine.
  This was so the host name of the HTTP Request would match the name on the
  certificate that is provided on our Splunk container.
  1. Use the provided `cacert.pem` file on the Splunk docker container for
  the certificate, and write it to the local machine
     ```
      docker exec -it splunk sudo cat /opt/splunk/etc/auth/cacert.pem > myLocalSplunkCertificate.pem
     ```
  1. Set a variable in the local environment named
  [SECRETLESS_HTTP_CA_BUNDLE](https://docs.conjur.org/latest/en/Content/References/connectors/scl_handlers-https.htm?TocPath=Fundamentals%7CSecretless%20Pattern%7CService%20Connectors%7CHTTP%7C_____0)
  and set it to the path where `myLocalSpunkCertificate.pem` was on the local
  machine.
  1. Run Secretless
     ```
      ./dist/darwin/amd64/secretless-broker \
      -f examples/generic_connector_configs/splunk_secretless.yml
     ```
  1. On another terminal window, make a request to Splunk using Secretless
     ```
      http_proxy=localhost:8081 curl -k -X GET \
      SplunkServerDefaultCert:8089/services/apps/local
     ```

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

> This configuration uses v2020-03-02. of the Stripe API. A
[`Stripe-Version` header](https://stripe.com/docs/api/versioning)
can be added to use this version.

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
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Get the Stripe test [API Key](https://dashboard.stripe.com/apikeys)
  1. Store the token from your request in your local credential manager so that
  it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/stripe_secretless.yml
     ```
  1. On another terminal window, make a request to Stripe using Secretless
     ```
      http_proxy=localhost:{secretless-server} curl api.stripe.com/v1/charges
     ```

</details>

___

### Tableau API

This example can be used to interact with
[Tableau's API](https://help.tableau.com/current/api/rest_api/en-us/REST/rest_api.htm).

The configuraton file for the Tableau API can be found at
[tableau_secretless.yml](./tableau_secretless.yml).

> This configuration uses v3.6 of the Tableau API.

#### How to use this connector

* Create an account for Tableau Online
* Make a request to Tableu's API to get an
[`X-Tableau-Auth`](https://help.tableau.com/current/api/rest_api/en-us/REST/rest_api_concepts_auth.htm)
token.
* Run Secretless with the supplied configuration(s)
* Query the API using localhost:8071 curl {data} {Tableau Endpoint URl}

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. Create an account on
  [Tableau Online](https://www.tableau.com/products/cloud-bi#form)
  1. [Make a Post Request](https://help.tableau.com/current/api/rest_api/en-us/REST/rest_api_get_started_tutorial_part_1.htm#step-1-sign-in-to-your-server-with-rest)
  to Tableau Online's API using the provided credentials to secure a
  `X-Tableau-Auth`
  1. Store the token from your request in your local credential manager so that
  it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/tableau_secretless.yml
     ```
  1. On another terminal window, make a request to Tableau using Secretless
     ```
      http_proxy=localhost:8071 curl -d {data} {Tableau Endpoint URL}
     ```

</details>

___

### Twilio API

This example can be used to interact with
[Twilio's Rest API](https://www.twilio.com/docs/usage/api).

The configuraton file for the Tableau API can be found at
[twilio_secretless.yml](./twilio_secretless.yml).

> This configuration uses version 2010-04-01 of the Twilio API.

#### How to use this connector

* [Create an account](https://www.twilio.com/try-twilio) for Twilio
* Edit the supplied configuration to get your
  [Account SID and Auth Token](https://support.twilio.com/hc/en-us/articles/223136027-Auth-Tokens-and-How-to-Change-Them)
  from the correct provider/path.
* Run Secretless with the supplied configuration(s)
* Query the API using localhost:9030 curl {data} {Twilio Endpoint URL}

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  1. [Create an account](https://www.twilio.com/try-twilio) for Twilio
  1. Get your [Account SID and Auth Token](https://www.twilio.com/console)
     from the Twilio console
  1. Store the Account SID and Auth Token from your request in your local
     credential manager so that it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
       ./dist/darwin/amd64/secretless-broker \
       -f examples/generic_connector_configs/tableau_secretless.yml
     ```
  1. On another terminal window, make a request to Tableau using Secretless
     ```
     http_proxy=localhost:9030 \
     curl -X POST api.twilio.com/2010-04-01/Accounts/{acc-sid}/Messages.json \
      --data-urlencode "Body=Hello from Secretless!" \
      --data-urlencode "From=+1{your-phone}" \
      --data-urlencode "To=+1{to-phone}"
     ```

</details>

___

### Twitter API

This example can be used to interact with
[Twitter's API](https://developer.twitter.com/en/docs).

The configuration file for the Twitter API can be found at
[twitter_secretless.yml](./twitter_secretless.yml).

> This configuration uses [v7](https://developer.twitter.com/en/docs/ads/general/overview/versions)
> of the Twitter API.

#### How to use this connector

* Edit the supplied service configuration to get your
[OAuth token](https://developer.twitter.com/en/docs/basics/authentication/oauth-2-0/bearer-tokens)
or [OAuth Consumer Key, Consumer Secret, Token Key and Token Secret](https://developer.twitter.com/en/docs/basics/authentication/oauth-1-0a/obtaining-user-access-tokens)
* Run Secretless with the supplied configuration(s)
* Query the API using `http_proxy=localhost:8051 curl api.twitter.com/{Request}`

#### Example Usage
<details>
  <summary><b>Example setup to try this out locally...</b></summary>

  **OAuth1:**
  1. Get your
     [Twitter Consumer Key and Consumer Secret](https://developer.twitter.com/en/apps)
     and generate a Token Key and Secret.
  1. Store the token from your request in your local credential manager so that
     it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
      ./dist/darwin/amd64/secretless-broker \
        -f examples/generic_connector_configs/twitter_secretless.yml
     ```
  1. On another terminal window, make a request to Twitter using Secretless
     ```
      http_proxy=localhost:8061 \
      curl "api.twitter.com/1.1/statuses/update.json?status=hello%20world"
     ```

  **OAuth2:**
  1. Get your
  [Twitter Consumer Key and Consumer Secret](https://developer.twitter.com/en/apps)
  1. Get an
  [OAuth token](https://developer.twitter.com/en/docs/basics/authentication/oauth-2-0/bearer-tokens)
  from Twitter through cURL
     ```
      curl -u 'API key:API secret key' \
      --data 'grant_type=client_credentials' \
      'https://api.twitter.com/oauth2/token'
     ```
  1. Store the token from your request in your local credential manager so that
  it may be retrieved in your `secretless.yml`
  1. Run Secretless locally
     ```
      ./dist/darwin/amd64/secretless-broker \
      -f examples/generic_connector_configs/twitter_secretless.yml
     ```
  1. On another terminal window, make a request to Twitter using Secretless
     ```
      http_proxy=localhost:8051 \
      curl "api.twitter.com/1.1/followers/ids.json?screen_name=twitterdev"
     ```

</details>

## Contributing

Do you have an HTTP service that you use? Can you write a Secretless generic
connector config for it? **Add the sample config to this folder and list it in
the table above!** Others may find your connector config useful, too - [send us
a PR](https://github.com/cyberark/community/blob/master/CONTRIBUTING.md#contribution-workflow)!
