---
title: Secretless
id: providers
layout: docs
description: Secretless Documentation
permalink: docs/reference/providers
---

# Credential Providers 

Credential Providers interact with a credential source to deliver secrets needed for authentication
to Secretless Listeners and Handlers. The Secretless broker comes built-in with several different
Credential Providers, making it easy to use with your existing workflows regardless of your current
secrets management toolset.

We currently support the following credential providers/vaults:

<div id="provider-tabs">
  <ul>
    <li><a href="#tabs-conjur-provider">CyberArk Conjur</a></li>
    <li><a href="#tabs-hashicorp-vault-provider">HashiCorp Vault</a></li>
    <li><a href="#tabs-file-provider">File Provider</a></li>
    <li><a href="#tabs-environment-variable-provider">Environment Variable</a></li>
    <li><a href="#tabs-literal-value-provider">Literal Value</a></li>
    <li><a href="#tabs-keychain-provider">Keychain</a></li>
  </ul>

  <div id="tabs-conjur-provider">
    <p>Conjur (<code>conjur</code>) provider allows use of <a href="https://www.conjur.org">CyberArk Conjur</a> for fetching secrets.</p>

    <p>Example:</p>
    <pre>
    ...
      credentials:
        - name: accessToken
          provider: conjur
          id: path/to/the/token
    ...
    </pre>
  </div>

  <div id="tabs-hashicorp-vault-provider">
    <p>Vault (<code>vault</code>) provider allows use of <a href="https://www.vaultproject.io/">HashiCorp Vault</a> for fetching secrets.</p>

    <p>Example:</p>
    <pre>
    ...
      credentials:
        - name: accessToken
          provider: vault
          id: path/to/the/token
    ...
    </pre>
  </div>

  <div id="tabs-file-provider">
    <p>File (<code>file</code>) provider allows you to use a file available to the Secretless process
    and/or container as sources of credentials.</p>

    <p>Example:</p>
    <pre>
    ...
      credentials:
        - name: rsa
          provider: file
          id: /path/to/file
    ...
    </pre>
  </div>

  <div id="tabs-environment-variable-provider">
    <p>Environment (<code>env</code>) provider allows use of environment variables as
    source of credentials.</p>

    <p>Example:</p>
    <pre>
    ...
      credentials:
        - name: accessToken
          provider: env
          id: ACCESS_TOKEN
    ...
    </pre>
  </div>

  <div id="tabs-literal-value-provider">
    <p>Literal (<code>literal</code>) provider allows use of hard-coded values as
    credential sources.</p>

    <p><em>Note: This type of secrets inclusion is highly likely to be much less secure
    versus other available providers so please use care when choosing this as your secrets
    source.</em></p>

    <p>Example:</p>
    <pre>
    ...
      credentials:
        - name: accessToken
          provider: literal
          id: supersecretaccesstoken
    ...
    </pre>
  </div>

  <div id="tabs-keychain-provider">
    <p>Keychain (<code>keychain</code>) provider allows use of your OS-level keychain as the
    credentials provider.</p>

    <p><em>Note: This provider currently only works on Mac OS at the time and only when building
    from source so it should be avoided unless you are a developer working on the source code.
    There are plans to integrate all major OS keychains into this provider in a future release.</em></p>

    <p>Example:</p>
    <pre>
    ...
      credentials:
        - name: rsa
          provider: keychain
          id: servicename#accountname
    ...
    </pre>
  </div>
</div>

<script>
  $( function() {
    $( "#provider-tabs" ).tabs();
  } );
</script>
