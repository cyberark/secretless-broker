# Generic HTTP Connector Example Configurations

The [generic HTTP connector](../../internal/plugin/connectors/http/generic/README.md)
enables using Secretless with a wide array of HTTP-based services _without
having to write new Secretless connectors_. Instead, you can modify your
Secretless configuration to specify the header structure the HTTP service
requires to authenticate.

## Sample Configurations

> Note: The following examples use the [Keychain provider](https://docs.cyberark.com/Product-Doc/OnlineHelp/AAM-DAP/11.3/en/Content/References/providers/scl_keychain.htm?TocPath=Fundamentals%7CSecretless%20Pattern%7CSecret%20Providers%7C_____5).
> Replace the service prefix `service#` with an appropriate service
> or use a different provider as needed.

|HTTP Service|Config File|Example Usage|
|---|---|---|
|OAuth 2.0 APIs|[oauth2_secretless.yml](./oauth2_secretless.yml)|This configuration acts as a generic OAuth2 connector. It can be used with any service that requires a Bearer token Authorization header.<ul><li>Edit the supplied service configuration to get your OAuth token</li><li>Run secretless with the supplied configuration(s)</li><li>Query the API using `http_proxy=localhost:8071 curl <Your OAuth2 API Endpoint URL>/{Request}`</li></ul>
|[GitHub API](https://developer.github.com/v3/)|[github_secretless.yml](./github_secretless.yml)|<ul><li>Edit the supplied configuration to get your [GitHub OAuth token](https://developer.github.com/v3/#oauth2-token-sent-in-a-header) from the correct provider/path.</li><li>Run Secretless with the supplied configuration</li><li>Query the GitHub API using `http_proxy=localhost:8081 curl api.github.com/{request}`</li></ul>|
|[Elasticsearch API](https://www.elastic.co/guide/en/elasticsearch/reference/current)|[elasticsearch_secretless.yml](./elasticsearch_secretless.yml)|<ul><li>Edit the supplied configuration to get your Elasticsearch [Api Key](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html) or [OAuth token](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-get-token.html)</li><li>Run secretless with the supplied configuration(s)</li><li>Query the Elasticsearch API using `http_proxy=localhost:9020 curl <Elasticsearch Endpoint URL>/{Request}`</li></ul>
|[Slack Web API](https://api.slack.com/apis)|[slack_secretless.yml](./slack_secretless.yml)|<ul><li>Edit the supplied configuration to get your Slack [OAuth token](https://api.slack.com/legacy/oauth#flow)</li><li>Run secretless with the supplied configuration(s)</li><li>Query the Slack API using `http_proxy=localhost:9030 curl -d {data} <Slack Endpoint URL>` or `http_proxy=localhost:9040 curl -d {data} <Slack Endpoint URL>` depending on if your endpoint requires JSON or URL encoded requests</li></ul>

## Contributing

Do you have an HTTP service that you use? Can you write a Secretless generic
connector config for it? **Add the sample config to this folder and list it in
the table above!** Others may find your connector config useful, too - [send us
a PR](https://github.com/cyberark/community/blob/master/CONTRIBUTING.md#contribution-workflow)!
