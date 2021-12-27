package main

import (
	"net/http"
	"strings"
	"time"
)

type roundTripper struct {
	*http.Transport
	UA       string
	Header   map[string]string
	MinDelay time.Duration // Minimum amount of time between requests.
	lastHit  time.Time     // Time of last request.
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(rt.UA) > 0 {
		req.Header.Set("User-Agent", rt.UA)
	}
	if len(rt.Header) > 0 {
		for key, value := range rt.Header {
			if len(req.Header.Get(key)) == 0 {
				req.Header.Set(key, value)
			}
		}
	}
	if strings.Contains(req.URL.RawQuery, " ") {
		req.URL.RawQuery = strings.Replace(req.URL.RawQuery, " ", "%20", -1)
	}
	if rt.MinDelay > 0 {
		diff := time.Now().Sub(rt.lastHit)
		if diff < rt.MinDelay {
			time.Sleep(diff)
		}
		rt.lastHit = time.Now()
	}
	res, err := rt.Transport.RoundTrip(req)
	if err != nil {
		return res, err
	}
	return res, err
}
