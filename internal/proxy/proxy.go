package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target string) (http.Handler, error) {

	urlParsed, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("could not parse url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(urlParsed)

	return proxy, nil

}
