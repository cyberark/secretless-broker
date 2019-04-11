# Manual tests for CRD

The tests are used to verify the usage of the Secretless Configuration CRDs with Secretless.

The tests proceed as follows:
1. Deploy Echo-Server, and Secretless Sidecar (deploys the CRDs using privileged ServiceAccount)
2. Create v1 Configuration CRD instance expected by Secretless, which sets up an HTTP proxy using the HTTP handler on port 8000
3. Make a call to the Echo-Server on port 8080 using Secretless as an HTTP proxy and assert the existence and value of credentials in the response headers
4. Update Configuration CRD instance to v1 expected by Secretless, which updates the values of the credentials
5. Repeat step 3, this time asserting on the new value of the credentials in the response headers

## Prerequisites

+ `kubectl` installed and already logged onto a Kubernetes cluster.
+ export `SECRETLESS_IMAGE` to point to the Secretless container image under test, e.g. `cyebrark/secretless-broker:latest`. This image must be available to be pulled by the nodes in your Kubernetes cluster.

## Usage

Run the tests with:
```bash
./deploy
```

Expected output:
```
cleaning up previous deployments
cleaned

secretless sidecar deploying CRD
deployed

waiting for CRD to be ready
.
ready

waiting for pod to be ready
......
ready

[TEST] create configuration object

applying manifest
waiting for pod to be ready

ready

testing
test passed ✔

[TEST] update configuration object

applying manifest
waiting for pod to be ready

ready

testing
test passed ✔

cleaning up previous deployments
cleaned

```
