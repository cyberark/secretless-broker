## conjur-oss secretless demo

Ensure you have:

+ docker
+ helm (CLI and already installed inside your Kubernetes cluster)
+ kubectl

1. If you have Conjur and your application database set up already  configure your own environment variables in `./pre-env.sh`. If not, jump to the README in requirements which goes through installing Conjur and MySQL.
1. Configure your own environment variables in `./env.sh`
1. Run through the numbered steps from `00` to `01`
1. Notice that in the last step the mysql service connector makes it possible for the application to connect to the database without handling the credentials`
1. When you're done, clean up by running `./-00-clean.sh`
