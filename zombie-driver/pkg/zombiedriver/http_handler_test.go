package zombiedriver_test

import (
	"context"
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

	"github.com/heetch/georgysavva-technical-test/zombie-driver/pkg/zombiedriver"
	"github.com/heetch/georgysavva-technical-test/zombie-driver/pkg/zombiedriver/mocks"
)

func TestHTTP_GetLocations(t *testing.T) {
	t.Parallel()
	ts, serviceMock := setupHTTPServer()
	defer ts.Close()
	serverURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	serviceMock.On(
		"GetDriver",
		mock.MatchedBy(func(_ context.Context) bool { return true }), // match anything of type context.Context
		defaultDriverID,
	).Return(&zombiedriver.Driver{ID: defaultDriverID, IsZombie: true}, nil)

	reqURL := serverURL.ResolveReference(&url.URL{Path: fmt.Sprintf("drivers/%s", defaultDriverID)})
	response, err := http.Get(reqURL.String())
	require.NoError(t, err)
	defer response.Body.Close()
	responseBytes, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)
	responseData := string(responseBytes)

	expectedResponseData := `
	{
		"id": "foo",
		"zombie": true
	}`
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	assert.JSONEq(t, expectedResponseData, responseData)
	serviceMock.AssertExpectations(t)
}

func setupHTTPServer() (*httptest.Server, *mocks.Service) {
	logger := log.New()
	logger.Level = log.ErrorLevel
	serviceMock := &mocks.Service{}
	hh := zombiedriver.MakeHTTPHandler(serviceMock, logger)
	ts := httptest.NewServer(hh)
	return ts, serviceMock
}
