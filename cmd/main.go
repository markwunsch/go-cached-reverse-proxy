package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/markwunsch/go-cached-reverse-proxy/internal/cache"
	"github.com/markwunsch/go-cached-reverse-proxy/internal/transport"
)

var (
	upstream     string
	upstreamURL  *url.URL
	timeout      time.Duration
	keepAlive    time.Duration
	maxIdleConns int
	err          error
)

func init() {
	flag.StringVar(&upstream, "UPSTREAM_URL", "", "URL for destination api in format https://google.com")
	flag.DurationVar(&timeout, "UPSTREAM_TIMEOUT", 120*time.Second, "Timeout duration for upstream api call")
	flag.DurationVar(&keepAlive, "UPSTREAM_KEEPALIVE", 30*time.Second, "KeepAlive duration for upstream api call")
	flag.IntVar(&maxIdleConns, "UPSTREAM_MAX_IDLE_CONNS", 100, "Max idle connections for upstream api calls")
}

func main() {
	flag.Parse()
	upstreamURL, err = url.Parse(upstream)
	if err != nil {
		log.Fatalf("failed to parse upstream URL: %v", err)
	}

	// local cache implementation
	cacher := cache.NewLocal()

	cachedRoundtrip := transport.NewCachedRoundtrip(
		http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: keepAlive,
			}).DialContext,
			MaxIdleConns:          maxIdleConns,
			IdleConnTimeout:       timeout,
			ExpectContinueTimeout: keepAlive,
		},
		cacher,
		upstreamURL.Host)

	p := httputil.NewSingleHostReverseProxy(upstreamURL)

	// substitute the upstream url in the request
	p.Director = transformRequest
	// use custom transporter that performs caching
	p.Transport = cachedRoundtrip
	// cache response from upstream api
	p.ModifyResponse = cachedRoundtrip.CacheResponse

	log.Fatal(http.ListenAndServe(":8080", p))
}

func transformRequest(req *http.Request) {
	req.URL.Host = upstreamURL.Host
	req.Host = upstreamURL.Host
	req.URL.Scheme = upstreamURL.Scheme
}
