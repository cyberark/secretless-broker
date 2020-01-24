# [TODO: Add connector name] Connector

TODO: Add description of connector

## Requirements

TODO: Add any requirements on the target service, its configuration, etc
that apply to your service connector here.

## Known Limitations

TODO: Add known limitations here

## Using the [TODO: Add connector name] connector

### Configuration

TODO: Add the required credentials and custom config required by your connector

#### Example Configurations

TODO: Add some sample Secretless configurations for your connector, or
edit the example below:


```yaml
services:
  my_example_service:
    connector: # Add your custom connector ID
    listenOn: tcp://0.0.0.0:8080
    credentials:
      credential1:
        from: conjur
        get: /path/to/credential1  # the id of your cred within conjur
    config:
      # Add your custom config here
```

### Troubleshooting

TODO: add some tips that end users can use to troubleshoot your connector, in
case it is not working as expected for them.

## Developer Documentation

TODO: If desired, add any special instructions for development on the connector
here.
