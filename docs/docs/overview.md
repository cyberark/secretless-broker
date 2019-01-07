---
title: Overview
id: overview
description: Secretless Broker Documentation
permalink: docs/overview.html
redirect_to: https://docs.secretless.io/Latest/en/Content/Resources/_TopNav/cc_Home.htm
---

The Secretless Broker is a connection broker that relieves client applications of
the need to directly handle secrets. When an application requires access to a
Target Service such as a database, web service, SSH connection, or any other
TCP-based service, rather than connect to the Target Service directly it can
connect to the local Secretless Broker <em>without credentials</em>. Secretless
Broker can be configured to retrieve credentials for each connection from any of
several credential stores and inject the credentials into the connection request.
Once the connection is made, Secretless Broker seamlessly streams the connection
between the client and the Target Service. The Secretless Broker can coordinate
connections to multiple Target Services in parallel.

In this section of the documentation, we will provide the motivation for
[why you should use the Secretless Broker](/docs/overview/why_secretless.html),
talk about [how it works](/docs/overview/how_it_works.html), and define some
[key terms](/docs/overview/key_terms.html). To find out more about currently
supported Target Services, please take a look at our [reference](/docs/reference.html).
