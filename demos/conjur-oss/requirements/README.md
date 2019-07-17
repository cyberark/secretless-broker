## conjur-oss secretless demo requirementss

Ensure you have:

+ docker
+ helm (CLI and already installed inside your Kubernetes cluster)
+ kubectl

Conjur:
  1. Configure your own environment variables in `./env.sh`
  1. Run through the numbered steps from `00`

MySQL:
  1. Configure your own environment variables in `./env.sh`
  1. Run through the numbered steps from `00`

NOTE: Both Conjur and MySQL have `-00-clean.sh` which carries out cleanup

Generate prerequisites environment variables for demo:
  1. Run `./gen-env.sh` from this directory.

Proceed with the README in the parents folder