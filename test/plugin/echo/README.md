This service is a simple echo server.

The server listens for a special http sequence ('\r\n') and kills the connection
once two of those are in a row and returns the content that was sent to it during that
time.


NOTE: This server does _not_ conform to HTTP response spec so it cannot be used with
regular `http.Get`!

Usage:

Run the server by calling
```
go run main.go
```

Send it a request as:
```
curl -A "UserAgentString" http://localhost:6174
```

To receive:
```
GET / HTTP/1.1
Host: localhost:6174
User-Agent: UserAgentString
Accept: */*
```
