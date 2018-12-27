---
title: Health Check
id: healthcheck
layout: docs
description: Secretless Broker Documentation
permalink: docs/reference/health-check/overview.html
---

Secretless broker exposes two endpoints that can be used for readiness and liveliness checks:

- `http://<host>:5335/ready` which will indicate if the broker has loaded a valid configuration.
- `http://<host>:5335/live` which will indicate if the broker has listeners activated.

If there are failures, the service will return a `503` status or a `200` if the service indicates that
the broker is ready/live.

Note: If Secretless is not provided with a configuration (e.g. it is not listening to anything),
the live endpoint will also return 503.

You can manually check the status with these endpoints by using `curl`:
```
$ # Start Secretless Broker in another terminal on the same machine

$ curl -i localhost:5335/ready
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 27 Dec 2018 17:12:07 GMT
Content-Length: 3

{}
```

If you would like to retrieve the full informational JSON that includes details on
which checks failed and with what error, you can add the `?full=1` query parameter
to the end of either of the available endpoints:
```
$ # Start Secretless Broker in another terminal on the same machine

$ curl -i localhost:5335/ready?full=1
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Thu, 27 Dec 2018 17:13:22 GMT
Content-Length: 45

{
    "listening": "OK",
    "ready": "OK"
}
```
