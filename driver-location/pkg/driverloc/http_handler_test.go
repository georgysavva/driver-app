package driverloc_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc"
	"github.com/heetch/georgysavva-technical-test/driver-location/pkg/driverloc/mocks"

	log "github.com/sirupsen/logrus"
)

func TestHTTP_GetLocations(t *testing.T) {
	t.Parallel()
	ts, serviceMock := setupHTTPServer()
	defer ts.Close()

	minutesArg := 5
	serviceMock.On(
		"GetLocations",
		mock.MatchedBy(func(_ context.Context) bool { return true }), // match anything of type context.Context
		defaultDriverID,
		time.Duration(minutesArg)*time.Minute,
	).Return([]*driverloc.Location{
		{
			Coordinates: &driverloc.Coordinates{
				Latitude:  48.864193,
				Longitude: 2.350498,
			},
			Time: time.Date(2018, 04, 05, 22, 36, 16, 00, time.UTC),
		},
		{
			Coordinates: &driverloc.Coordinates{
				Latitude:  48.863921,
				Longitude: 2.349211,
			},
			Time: time.Date(2018, 04, 05, 22, 36, 21, 00, time.UTC),
		},
	}, nil)

	queryParams := map[string]string{"minutes": strconv.Itoa(minutesArg)}
	response, responseData := callGetLocationsEndpoint(t, ts, queryParams)
	expectedResponseData := `
	[{
		"latitude": 48.864193,
		"longitude": 2.350498,
		"updated_at": "2018-04-05T22:36:16Z"
	}, {
		"latitude": 48.863921,
		"longitude": 2.349211,
		"updated_at": "2018-04-05T22:36:21Z"
	}]`
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	assert.JSONEq(t, expectedResponseData, responseData)
}

func TestHTTP_GetLocations_MinutesParamMissing(t *testing.T) {
	t.Parallel()
	ts, _ := setupHTTPServer()
	defer ts.Close()

	queryParams := map[string]string{}
	response, responseData := callGetLocationsEndpoint(t, ts, queryParams)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Equal(t, "'minutes' query param is missing\n", responseData)
}

func TestHTTP_GetLocations_MinutesParamNotANumber(t *testing.T) {
	t.Parallel()
	ts, _ := setupHTTPServer()
	defer ts.Close()

	queryParams := map[string]string{"minutes": "five"}
	response, responseData := callGetLocationsEndpoint(t, ts, queryParams)

	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Equal(t, "'minutes' query param must be a number\n", responseData)
}

func callGetLocationsEndpoint(t *testing.T, ts *httptest.Server, queryParams map[string]string) (*http.Response, string) {
	t.Helper()
	serverURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	query := url.Values{}
	for k, v := range queryParams {
		query.Set(k, v)
	}

	reqURL := serverURL.ResolveReference(&url.URL{
		Path:     fmt.Sprintf("drivers/%s/locations", defaultDriverID),
		RawQuery: query.Encode(),
	})

	resp, err := http.Get(reqURL.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	bodyText := string(bodyBytes)
	return resp, bodyText
}

func setupHTTPServer() (*httptest.Server, *mocks.GetterService) {
	logger := log.New()
	logger.Level = log.ErrorLevel
	serviceMock := &mocks.GetterService{}
	hh := driverloc.MakeHTTPHandler(serviceMock, logger)
	ts := httptest.NewServer(hh)
	return ts, serviceMock
}
