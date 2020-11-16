package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const httpUpstreamScheme = "http"

type Endpoint struct {
	Path   string         `yaml:"path"`
	Method string         `yaml:"method"`
	NSQ    *NSQProxyConf  `yaml:"nsq"`
	HTTP   *HTTPProxyConf `yaml:"http"`
}

type NSQProxyConf struct {
	Topic   string          `yaml:"topic"`
	Message *NSQMessageConf `yaml:"message"`
}

type NSQMessageConf struct {
	Command string `yaml:"command"`
}

type HTTPProxyConf struct {
	Host string `yaml:"host"`
}

func NewGateway(nsqFactory *NSQProxyFactory, endpoints []*Endpoint) (http.Handler, error) {
	router := mux.NewRouter()
	for _, endpoint := range endpoints {
		if endpoint.HTTP != nil && endpoint.NSQ != nil {
			return nil, errors.Errorf("endpoint must contain either nsq or http proxy configs, not both: %+v", endpoint)
		}
		if endpoint.HTTP == nil && endpoint.NSQ == nil {
			return nil, errors.Errorf("endpoint must contain either nsq or http proxy configs, not none: %+v", endpoint)
		}
		var proxyHandler http.Handler
		if endpoint.HTTP != nil {
			proxyConf := endpoint.HTTP
			if proxyConf.Host == "" {
				return nil, errors.Errorf("endpoint has an empty http host: %+v", endpoint)
			}
			targetURL := &url.URL{Scheme: httpUpstreamScheme, Host: proxyConf.Host}
			proxyHandler = httputil.NewSingleHostReverseProxy(targetURL)
		} else {
			proxyConf := endpoint.NSQ
			var err error
			if proxyHandler, err = nsqFactory.NewProxy(proxyConf); err != nil {
				return nil, errors.Wrapf(err, "can't initialize nsq proxy for endpoint: %+v", endpoint)
			}
		}

		router.Handle(endpoint.Path, proxyHandler).Methods(endpoint.Method)
	}
	return router, nil
}
