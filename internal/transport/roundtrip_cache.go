package transport

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/markwunsch/go-cached-reverse-proxy/internal/cache"
)

const CacheHit = `hit`

type CachedRoundrip struct {
	t     http.Transport
	cache cache.Manager
	host  string
}

func NewCachedRoundtrip(t http.Transport, c cache.Manager, h string) *CachedRoundrip {
	return &CachedRoundrip{
		cache: c,
		host:  h,
		t:     t,
	}
}

func (c *CachedRoundrip) CacheResponse(resp *http.Response) error {
	if resp.StatusCode != 200 || resp.Header.Get("x-cache") == CacheHit {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(`error reading response: %v`, err)
		return err
	}

	// mwunsch TODO: cache headers as well
	key := cacheKey(resp.Request.RequestURI)
	if err := c.cache.Put(key, string(body)); err != nil {
		return err
	}

	resp.Body.Close()
	resp.Body = io.NopCloser(strings.NewReader(string(body)))
	return nil
}

func (c *CachedRoundrip) RoundTrip(r *http.Request) (w *http.Response, err error) {
	var uri = r.URL.RequestURI()
	k := cacheKey(uri)

	b, err := c.cache.Get(k)
	if err == nil {
		log.Printf("cache hit")

		// mwunsch TODO: store and retrieve headers from cache
		h := make(http.Header)
		h.Set("x-cache", CacheHit)
		w = &http.Response{
			Request:    r,
			Body:       io.NopCloser(strings.NewReader(b)),
			Header:     h,
			Status:     "200 OK",
			StatusCode: 200,
		}
		return w, nil
	} else {
		log.Printf("error getting cached value: %v)", err)
	}

	w, err = c.t.RoundTrip(r)
	if err != nil {
		log.Printf("error returned during request: %v", err)
		return nil, err
	}

	return w, err
}

func cacheKey(uri string) string {
	return base64.URLEncoding.EncodeToString([]byte(uri))
}
