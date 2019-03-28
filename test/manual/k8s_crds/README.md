# Manual tests for CRD

The tests are used to verify the usage of the Secretless Configuration CRDs with Secretless.

The tests proceed as follows:
1. Deploy CRDs, and Echo-Server with Secretless Sidecar
2. Create v1 Configuration CRD instance expected by Secretless, which sets up an HTTP proxy using the HTTP handler on port 8000
3. Make a call to the Echo-Server on port 8080 using Secretless as an HTTP proxy and assert the existence and value of credentials in the response headers
4. Update Configuration CRD instance to v1 expected by Secretless, which updates the values of the credentials
5. Repeat step 3, this time asserting on the new value of the credentials in the response headers

## Prerequisites

+ `kubectl` installed and already logged onto a Kubernetes cluster.
+ working from the `default` namespace
+ `secretless-broker:latest` is present in the container store of your Kubernetes cluster.

## Usage

After you're done don't forget to clean up with:
```bash
./stop_deployment
```

Run the tests with:
```bash
./deploy
```

Expected output:
```
cleaning up previous deployments
cleaned

deploying CRDs
deployed

waiting for CRDs to be ready
ready

waiting for pod to be ready
.....
ready

[TEST] v1 CRD config

creating v1 CRD config
testing
test passed

[TEST] v2 CRD config

updating CRD config to v2
testing
test passed
```
