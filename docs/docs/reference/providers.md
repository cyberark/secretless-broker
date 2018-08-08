---
title: Credential Providers
id: providers
layout: docs
description: Secretless Documentation
permalink: docs/reference/providers
--- 

Credential Providers interact with a credential source to deliver secrets needed for authentication
to Secretless Listeners and Handlers. The Secretless broker comes built-in with several different
Credential Providers, making it easy to use with your existing workflows regardless of your current
secrets management toolset.

We currently support the following credential providers/vaults:
- [Conjur](/docs/reference/providers/conjur.html)
- [Environment](/docs/reference/providers/env.html)
- [File](/docs/reference/providers/file.html)
- [Keychain](/docs/reference/providers/keychain.html)
- [Literal](/docs/reference/providers/literal.html)
