
### Connecting to Secretless as an HTTPS proxy
#### Symptoms
- You see a CONNECT request in your Secretless logs that looks something like:

  ```
  2020/05/06 18:39:19 [DEBUG] Got request  httpbin.org:443 CONNECT //httpbin.org:443
  ```

- Sample client log output messages
  
  curl:
  ```
  *   Trying 127.0.0.1...
  * TCP_NODELAY set
  * Connected to 127.0.0.1 (127.0.0.1) port 62160 (#0)
  * Establish HTTP proxy tunnel to httpbin.org:443
  > CONNECT httpbin.org:443 HTTP/1.1
  > Host: httpbin.org:443
  > User-Agent: curl/7.54.0
  > Proxy-Connection: Keep-Alive
  > 
  < HTTP/1.1 405 Method Not Allowed
  < Content-Type: text/plain; charset=utf-8
  < X-Content-Type-Options: nosniff
  < Date: Wed, 06 May 2020 17:39:19 GMT
  < Content-Length: 26
  < 
  * Received HTTP code 405 from proxy after CONNECT
  * Closing connection 0
  curl: (56) Received HTTP code 405 from proxy after CONNECT
  ```

  Go client:
  ```
  Get https://httpbin.org/anything: Method Not Allowed
  ```

#### Known Causes
This type of error occurs when the client attempts to use Secretless as an HTTPS proxy. 
Secretless can only act as an HTTP proxy.
This error can happen for a few reasons:
1. An explicit attempt to use Secretless as an HTTPS proxy
1. Providing the client an HTTPS target when intending to proxy the connection through Secretless might result in the client attempting to use Secretless as an HTTPS proxy, as is the case with Go's standard library HTTP client.

#### Resolution
- Ensure that target of your request is HTTP only e.g. http://httpbin.org.
- Secretless does not support HTTPS between the client and Secretless, it does support it between Secretless and the target. Do not use Secretless as an HTTPS proxy. 
- To make connection between Secretless and the target an HTTPS connection you must set "forceSSL: true" on the Secretless service connector config. 


### HTTPS certificate verification failure when forceSSL is set
#### Symptoms
- You see an x509 certificate error in your Secretless logs that looks something like:

  ```
  2020/05/06 18:39:19 [DEBUG] Got request / self-signed.badssl.com GET http://self-signed.badssl.com/
  2020/05/06 18:39:19 [DEBUG] Using connector 'test' for request http://self-signed.badssl.com/
  2020/05/06 18:39:20 [DEBUG] Error: x509: certificate signed by unknown authority
  ```

- Sample client log output messages
  
  curl:
  ```
  * Rebuilt URL to: http://self-signed.badssl.com/
  *   Trying 127.0.0.1...
  * TCP_NODELAY set
  * Connected to 127.0.0.1 (127.0.0.1) port 62165 (#0)
  > GET http://self-signed.badssl.com/ HTTP/1.1
  > Host: self-signed.badssl.com
  > User-Agent: curl/7.54.0
  > Accept: */*
  > Proxy-Connection: Keep-Alive
  > 
  < HTTP/1.1 503 Service Unavailable
  < Content-Type: text/plain; charset=utf-8
  < X-Content-Type-Options: nosniff
  < Date: Wed, 06 May 2020 17:39:20 GMT
  < Content-Length: 46
  < 
  { [46 bytes data]
  * Connection #0 to host 127.0.0.1 left intact
  x509: certificate signed by unknown authority
  ```

  Go client:
  ```
  Status:
  503 Service Unavailable
  
  Body: 
  x509: certificate is valid for *.badssl.com, badssl.com, not wrong.host.badssl.com
  ```

#### Known Causes
This type of error occurs when the client attempts to connect to a target with a self-signed certificate, and there is some failure on verification. Secretless verifies all HTTPS connections to the target.

There are several reasons why verification might fail including:
1. The signer of the target's certificate is not a trusted CA
1. The target's certificate is expired or is not yet valid
1. The target's certificate is not valid for the host

#### Resolution

This type of error can be broken into 2 categories.
1. The signer of the target's certificate is not a trusted CA
1. The rest

For the rest (2), you must ensure that the target's certificate is valid for the target. 

For (1) you will need to ensure that Secretless is aware of the root certificate authority (CA) it should use to verify the server certificates when proxying requests. To do this, ensure that the environment variable **SECRETLESS_HTTP_CA_BUNDLE** is set. **SECRETLESS_HTTP_CA_BUNDLE** is a path to the bundle of CA certificates that are appended to the certificate pool that Secretless uses for server certificate verification of all HTTP service connectors.
