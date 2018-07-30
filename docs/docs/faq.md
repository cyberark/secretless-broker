---
title: Secretless
id: faq
layout: docs
description: Secretless Documentation
permalink: docs/faq
---

<p class="card-heading">FAQ</p>

<div class="container-fluid" id="faq-list">
  <ul>
    <li><a href="#faq-why-secretless">Why is it called "Secretless" when there are still secrets being managed?</a></li>
    <li><a href="#faq-vault-not-supported">The vault I use is not supported! Can I still use Secretless?</a></li>
    <li><a href="#faq-service-not-supported">The service I use is not supported! Can I still use Secretless?</a></li>
  </ul>
</div>

<div class="faq" id="faq-why-secretless">
  Why is it called "Secretless" when there are still secrets being managed?
</div>

<p>Secrets <em>are</em> still being managed - but <strong>not by your applications</strong>.
  Which is huge! Before Secretless, the state of the art for secrets management was to store
  your secrets in a vault and update your applications to retrieve them from the vault.
  You could do this by updating your source code to interact directly with the vault
  API, or you could use a tool like <a href="https://cyberark.github.io/summon">Summon</a>
  to abstract away the API interaction and inject the secret values into your application's
  environment at runtime.</p>

<p>But even if you are following best practices and storing your secrets in a vault,
  regardless of how you set up your apps to retrieve the secrets you still have to:</p>
  <ul>
    <li>Securely handle retrieved secrets within app</li>
    <li>Resiliently handle secret rotations</li>
  </ul>

<p>Using the Secretless broker allows you to <em>remove consideration of secrets from
  your applications</em>. Once you use Secretless, your apps only have to worry about
  connecting to target services via a local socket or TCP connection <em>without providing
  credentials</em>, greatly simplifying the path to writing secure applications.</p>

<div class="faq" id="faq-vault-not-supported">
  The vault I use is not supported! Can I still use Secretless?
</div>

<p>For info on currently supported vaults, please see our
  <a href="/providers.html">Credential Providers</a> reference page.</p>

<p>If the vault you would like to use is not currently supported, please check our
  <a href="https://github.com/conjurinc/secretless/issues">Github issues</a> to see
  if we already have plans to support it. If not, please open a new issue with your
  request. Secretless is also <a href="community.html">open to contributions</a>
  from the community, and we plan to standardize the Provider API to make it easy
  to contribute a Credential Provider.</p>

<div class="faq" id="faq-service-not-supported">
  The service I use is not supported! Can I still use Secretless?
</div>

<p>For info on currently supported services, please see our <a href="/reference.html">reference</a>.</p>

<p>If the service you would like to use is not currently supported, please check our
  <a href="https://github.com/conjurinc/secretless/issues">Github issues</a> to see if
  we already have plans to support it. If not, please open a new issue with your request.
  Secretless is also <a href="community.html">open to contributions</a> from the community;
  please see our <a href="/generated/pkg_secretless_plugin_v1.html">plugin reference</a>
  for guidance on implementing new Listeners or Handlers (to enable Secretless to proxy
  connections to a new service).</p>
