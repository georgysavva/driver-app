package driverlochttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/georgysavva/driver-app/driver-location/pkg/clients/driverlochttp"
	"github.com/georgysavva/driver-app/driver-location/pkg/driverloc"
	"github.com/georgysavva/driver-app/driver-location/pkg/driverloc/mocks"
)

const defaultDriverID = "foo"

func TestClient_GetLocations(t *testing.T) {
	t.Parallel()
	ts, serviceMock := setupHTTPServer()
	defer ts.Close()

	timeInterval := 5 * time.Minute
	expected := []*driverloc.Location{
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
	}

	serviceMock.On(
		"GetLocations",
		mock.MatchedBy(func(_ context.Context) bool { return true }), // anything of type context.Context
		defaultDriverID, timeInterval,
	).Return(expected, nil)

	client, err := driverlochttp.NewClient(http.DefaultClient, ts.URL)
	require.NoError(t, err)
	actual, err := client.GetLocations(context.Background(), defaultDriverID, timeInterval)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
	serviceMock.AssertExpectations(t)
}

func setupHTTPServer() (*httptest.Server, *mocks.GetterService) {
	logger := log.New()
	logger.Level = log.ErrorLevel
	serviceMock := &mocks.GetterService{}
	hh := driverloc.MakeHTTPHandler(serviceMock, logger)
	ts := httptest.NewServer(hh)
	return ts, serviceMock
}
