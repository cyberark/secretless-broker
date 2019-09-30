/*
Package proxyservice takes a Secretless configuration and available plugins
and constructs the requires ProxyServices that Secretless will run.

To put this in context, the complete high-level flow is:

	1. Parse the `secretless.yml` into individual “service” configs,
	corresponding to each service entry in the yml.

	2. Identify http services that share a `listenOn`, because we know those
	will all be part of a single “http proxy service” that uses traffic routing
	within it to delegate to the subservice connectors.

	3. So now we have the http service's `listenOn` and all the “subservices”
	associated with it.

	4. Each of those subservices needs two things: a connector (which knows how
	to authenticate requests) and a way to get the current credentials at
	runtime.

	5. Now note the signature of the connector itself just looks like this:

		type Connector func(request *http.Request, secrets plugin.SecretsByID) error

	6. So it is the responsibility of the proxy service to actually fetch the
	credentials.  So what does the proxy service need for each of those
	subservices? Precisely this:

		type HTTPSubService struct {
		  connector http.Connector,
		  retrieveCredentials internal.CredentialsRetriever,
		}

	7. Putting all the together, here’s what we need to construct a new http
	proxy service:

		func NewProxyService(
			subservices []HTTPSubService,
			sharedListener net.Listener,
			logger log.Logger,
		) (internal.Service, error) {

One fine point that’s not be obvious: Each of the subservice connectors gets
created with its own custom logger, but those A. those aren’t accessible to the
proxy service itself and B. even if they were, we’d want to explicitly pass the
logger to be clear about this is a dependency of the proxy service itself and C.
the proxy service’s logger should have a different prefix than those of the more
specific subservices.  So that’s why we’re passing the logger too.

TODO: Add a principled explanation about how logging is working.
  Eg, when do we return errors, when is it fatal?  what do we log, everything?
*/
package proxyservice
