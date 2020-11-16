package gateway_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/heetch/georgysavva-technical-test/gateway/pkg/gateway"
	"github.com/heetch/georgysavva-technical-test/gateway/pkg/gateway/mocks"
)

func TestNSQProxy(t *testing.T) {
	t.Parallel()
	producerMock := &mocks.NSQProducer{}
	producerMock.On("Publish", "test-topic", mock.AnythingOfType("[]uint8")).Return(nil)
	logger := log.New()
	logger.SetLevel(log.ErrorLevel)
	proxyFactory := gateway.NewNSQProxyFactory(producerMock, logger)
	endpoints := []*gateway.Endpoint{
		{
			Path:   "/",
			Method: "POST",
			NSQ: &gateway.NSQProxyConf{
				Topic: "test-topic",
				Message: &gateway.NSQMessageConf{
					Command: "test_command",
				},
			},
		},
	}
	gatewayHandler, err := gateway.NewGateway(proxyFactory, endpoints)
	require.NoError(t, err)
	ts := httptest.NewServer(gatewayHandler)
	defer ts.Close()

	response, responseText := callEndpoint(t, ts, "/" /* endpointPath */)

	assert.Equal(t, response.StatusCode, http.StatusOK)
	assert.Equal(t, responseText, "OK\n")
	producerMock.AssertExpectations(t)
}

func TestHTTPProxy(t *testing.T) {
	t.Parallel()
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "OK")
		require.NoError(t, err)
	}))
	defer backendServer.Close()
	backendServerURL, err := url.Parse(backendServer.URL)
	require.NoError(t, err)

	endpoints := []*gateway.Endpoint{
		{
			Path:   "/",
			Method: "POST",
			HTTP: &gateway.HTTPProxyConf{
				Host: backendServerURL.Host,
			},
		},
	}
	gatewayHandler, err := gateway.NewGateway(nil /* nsqFactory */, endpoints)
	require.NoError(t, err)
	ts := httptest.NewServer(gatewayHandler)
	defer ts.Close()

	response, responseText := callEndpoint(t, ts, "/" /* endpointPath */)

	assert.Equal(t, response.StatusCode, http.StatusOK)
	assert.Equal(t, responseText, "OK\n")
}

func callEndpoint(t *testing.T, ts *httptest.Server, endpointPath string) (*http.Response, string) {
	t.Helper()
	serverURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	reqURL := serverURL.ResolveReference(&url.URL{Path: endpointPath})

	resp, err := http.Post(reqURL.String(), "application/json", nil /* body */)
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyText := string(bodyBytes)
	return resp, bodyText
}
