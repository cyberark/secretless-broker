## conjur-oss secretless demo

Ensure you have:

+ docker
+ helm (CLI and already installed inside your Kubernetes cluster)
+ kubectl

1. Configure your own environment variables in `./env.sh`
1. Run through the numbered steps from 00 to 04
1. Notice that in the last step the http service connector is used to inject credentials from conjur as basic auth headers i.e. `abcxyzusername:abcxyzpassword`
1. Clean up by running `./-00-clean.sh`
