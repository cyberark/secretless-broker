package troubleshooting

import (
	"net"
	"net/http"
	"net/url"
)

// ephemeralListenerOnPort creates a net.Listener (with a short deadline) on some given
// port. Note that passing in a port of "0" will result in a random port being used.
func ephemeralListenerOnPort(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return nil, err
	}

	// We generally don't want to wait forever for a connection to come in
	//err = listener.(*net.TCPListener).SetDeadline(time.Now().Add(10 * time.Second))
	//if err != nil {
	//	return nil, err
	//}

	return listener, nil
}

// proxyGet is a convenience method that makes an HTTP GET request using a proxy
func proxyGet(endpoint, proxy string) (*http.Response, error) {
	req, err := http.NewRequest(
		"GET",
		endpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{Proxy: func(req *http.Request) (proxyURL *url.URL, err error) {
		return url.Parse(proxy)
	}}
	client := &http.Client{Transport: transport}
	return client.Do(req)
}
