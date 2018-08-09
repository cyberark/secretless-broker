---
title: Quick Start
id: quick_start
layout: docs
description: Secretless Broker Documentation
permalink: docs/get_started/quick_start
---

Try out the Secretless Broker brokering a connection to a PostgreSQL database, an SSH connection,
or a connection to an HTTP service authenticating with basic auth.

<div id="quick-start-tabs">
  <ul>
    <li><a href="#tabs-demo-pg">PostgreSQL</a></li>
    <li><a href="#tabs-demo-ssh">SSH</a></li>
    <li><a href="#tabs-demo-http">HTTP</a></li>
  </ul>

  <div id="tabs-demo-pg">
    <ol>
      <li>
        <p>Download and run the Secretless Broker quick-start as a Docker container:</p>    
{% highlight shell %}
docker container run \
  --rm \
  -p 5432:5432 \
  -p 5454:5454 \
  cyberark/secretless-broker-quickstart
  {% endhighlight %}
      </li>

      <li>
        <p>Direct access to the PostgreSQL database is available over port
        <code>5432</code>. You can try querying some data, but you don't
        have the credentials required to connect (even if you know the
        username):</p>
{% highlight shell %}
psql \
  --host localhost \
  --port 5432 \
  --username secretless \
  -d quickstart \
  -c 'select * from counties;'
  {% endhighlight %}
      </li>

      <li>
        <p>The good news is that you don't need any credentials! Instead, you
        can connect to the password-protected PostgreSQL database via the
        Secretless Broker on port <code>5454</code>, <i>without knowing the
        password.</i> Give it a try:</p>
{% highlight shell %}
psql \
  --host localhost \
  --port 5454 \
  --username secretless \
  -d quickstart \
  -c 'select * from counties;'
  {% endhighlight %}
      </li>
    </ol>
  </div>

  <div id="tabs-demo-http">
    <ol>
      <li>
        <p>Download and run the Secretless Broker quick-start as a Docker container:</p>
{% highlight shell %}
docker container run \
  --rm \
  -p 8080:80 \
  -p 8081:8081 \
  cyberark/secretless-broker-quickstart
  {% endhighlight %}
      </li>

      <li>
        <p>The service we're trying to connect to is listening on port
        <code>8080</code>. If you try to access it, the service will inform
        you that you're unauthorized:</p>
        {% highlight shell %}curl -i localhost:8080{% endhighlight %}
      </li>

      <li>
        <p>Instead, you can make an authenticated HTTP request by proxying
        through the Secretless Broker on port <code>8081</code>. The Secretless Broker
        will inject the proper credentials into the request <i>without you
        needing to know what they are</i>. Give it a try:</p>
        {% highlight shell %}http_proxy=localhost:8081 curl -i localhost:8080{% endhighlight %}
      </li>
    </ol>
  </div>

  <div id="tabs-demo-ssh">
    <ol>
      <li>
        <p>Download and run the Secretless Broker quick-start as a Docker container:</p>
{% highlight shell %}
docker container run \
  --rm \
  -p 2221:22 \
  -p 2222:2222 \
  cyberark/secretless-broker-quickstart
  {% endhighlight %}
      </li>

      <li>
        <p>The default SSH service is exposed over port <code>2221</code>. You
        can try opening an SSH connection to the server, but you don't have
        the credentials to log in:</p>
        {% highlight shell %}ssh -p 2221 user@localhost{% endhighlight %}
      </li>

      <li>
        <p>The good news is that you don't need credentials! You can establish
        an SSH connection through the Secretless Broker on port
        <code>2222</code> <i>without any credentials</i>. Give it a try:</p>
        {% highlight shell %}ssh -p 2222 user@localhost{% endhighlight %}
      </li>
    </ol>
  </div>
</div>

<script>
  $( function() {
    $( "#quick-start-tabs" ).tabs();
  } );
</script>
