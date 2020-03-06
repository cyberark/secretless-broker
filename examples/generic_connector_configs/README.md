# Generic HTTP Connector Example Configurations

The [generic HTTP connector](../../internal/plugin/connectors/http/generic/README.md)
enables using Secretless with a wide array of HTTP-based services _without
having to write new Secretless connectors_. Instead, you can modify your
Secretless configuration to specify the header structure the HTTP service
requires to authenticate.

## Sample Configurations

|HTTP Service|Config File|Example Usage|
|---|---|---|
|[GitHub API](https://developer.github.com/v3/)|[github_secretless.yml](./github_secretless.yml)|<ul><li>Edit the supplied configuration to get your [GitHub OAuth token](https://developer.github.com/v3/#oauth2-token-sent-in-a-header) from the correct provider/path.</li><li>Run Secretless with the supplied configuration</li><li>Query the GitHub API using `http_proxy=localhost:8081 curl api.github.com/{request}`</li></ul>|

## Contributing

Do you have an HTTP service that you use? Can you write a Secretless generic
connector config for it? **Add the sample config to this folder and list it in
the table above!** Others may find your connector config useful, too - [send us
a PR](https://github.com/cyberark/community/blob/master/CONTRIBUTING.md#contribution-workflow)!
